// SPDX-License-Identifier: GPL-2.0
#include <stddef.h>
#include <linux/if_ether.h>
#include <string.h>
#include <linux/in.h>
#include <linux/bpf.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/icmp.h>
#include <linux/icmpv6.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include <stdint.h>

static __always_inline uint16_t csum_fold(uint32_t csum)
{
    /* The highest reasonable value for an IPv4 header
     * checksum requires two folds, so we just do that always.
     */
    csum = (csum & 0xffff) + (csum >> 16);
    csum = (csum & 0xffff) + (csum >> 16);
    return (uint16_t)~csum;
}

#define IP_DF (0x4000)
#define IPV4_HEADER_LENGTH 5
#define INNER_PKT_SIZE 64

SEC("icmp")
int bpf_lwt_icmp_too_big(struct __sk_buff *skb)
{
    if (skb->len < 1400) {
        return BPF_OK;
    }

    if (bpf_skb_change_tail(skb, INNER_PKT_SIZE, 0)) {
        return BPF_DROP;
    }

    void *data = (void*)(long)skb->data;
    void *data_end = (void*)(long)skb->data_end;

    if (skb->protocol == bpf_ntohs(ETH_P_IP)) {
        struct iphdr *inner_ipv4 = (struct iphdr*)data;
        if ((void*)(inner_ipv4 + 1) > data_end) {
            return BPF_DROP;
        }

        struct encap_hdr {
            struct iphdr ipv4;
            struct icmphdr icmp4;
        } encap;
        memset(&encap, 0, sizeof(struct encap_hdr));

        encap.ipv4.version = IPVERSION;
        encap.ipv4.ihl = IPV4_HEADER_LENGTH;
        encap.ipv4.tos = 0;
        encap.ipv4.tot_len = bpf_htons(sizeof(struct iphdr) + skb->len + sizeof(struct icmp6hdr));
        encap.ipv4.id = 0;
        encap.ipv4.frag_off = bpf_htons(IP_DF);
        encap.ipv4.ttl = IPDEFTTL;
        encap.ipv4.protocol = IPPROTO_ICMP;
        encap.ipv4.saddr = inner_ipv4->daddr;
        encap.ipv4.daddr = inner_ipv4->saddr;
        encap.ipv4.check = 0;
        encap.ipv4.check = csum_fold(bpf_csum_diff(0, 0, (__be32*)&encap.ipv4, sizeof(encap.ipv4), 0));

        encap.icmp4.type = ICMP_DEST_UNREACH;
        encap.icmp4.code = ICMP_FRAG_NEEDED;
        encap.icmp4.un.frag.mtu = bpf_htons(1350);
        encap.icmp4.checksum = 0;
        __u32 csum = bpf_csum_diff(0, 0, (__be32*)&encap.icmp4, sizeof(encap.icmp4), 0);
        if (data + INNER_PKT_SIZE > data_end) {
            return BPF_DROP;
        }
        encap.icmp4.checksum = csum_fold(bpf_csum_diff(0, 0, (__be32*)data, INNER_PKT_SIZE, csum));

        if (bpf_lwt_push_encap(skb, BPF_LWT_ENCAP_IP, &encap, sizeof(struct encap_hdr))) {
            return BPF_DROP;
        }
    } else if (skb->protocol == bpf_ntohs(ETH_P_IPV6)) {
        struct ipv6hdr *inner_ipv6 = (struct ipv6hdr*)data;
        if ((void*)(inner_ipv6 + 1) > data_end) {
            return BPF_DROP;
        }

        struct encap_hdr {
            struct ipv6hdr ipv6;
            struct icmp6hdr icmp6;
        } encap;
        memset(&encap, 0, sizeof(struct encap_hdr));

        encap.ipv6.version = 6;
        encap.ipv6.payload_len = bpf_htons(skb->len + sizeof(struct icmp6hdr));
        encap.ipv6.nexthdr = IPPROTO_ICMPV6;
        encap.ipv6.hop_limit = 0x40;
        // 交换内层 ipv4，ipv6 地址发回
        encap.ipv6.saddr = inner_ipv6->daddr;
        encap.ipv6.daddr = inner_ipv6->saddr;
        encap.icmp6.icmp6_type = ICMPV6_PKT_TOOBIG;
        encap.icmp6.icmp6_mtu = bpf_ntohl(1350);

        __u32 csum; 
        // 计算伪头 checksum
        struct psuedoHeader {
            struct in6_addr saddr;
            struct in6_addr daddr;
            __u32 payload_len;
            __u8 nexthdr[4];
        } ph;
        memset(&ph, 0, sizeof(ph));
        ph.saddr = encap.ipv6.saddr;
        ph.daddr = encap.ipv6.daddr;
        ph.payload_len = encap.ipv6.payload_len;
        ph.nexthdr[3] = IPPROTO_ICMPV6;
        csum = bpf_csum_diff(0, 0, (__be32*)&ph, sizeof(ph), 0);
        // 加 icmp6 头的 checksum
        csum = bpf_csum_diff(0, 0, (__be32*)&encap.icmp6, sizeof(encap.icmp6), csum);

        if (data + INNER_PKT_SIZE > data_end) {
            return BPF_DROP;
        }
        // 加 icmp6 包里面数据的 checksum 得到最终结果
        encap.icmp6.icmp6_cksum = csum_fold(bpf_csum_diff(0, 0, (__be32*)data, INNER_PKT_SIZE, csum));

        if (bpf_lwt_push_encap(skb, BPF_LWT_ENCAP_IP, &encap, sizeof(struct encap_hdr))) {
            return BPF_DROP;
        }
    } else {
        return BPF_OK;
    }

    // 添加 mac 头，否则 redirect 会失败
    struct ethhdr eth;
    memset(&eth, 0, sizeof(eth));
    eth.h_proto = skb->protocol;
    // 目的 mac 地址不能全为 0
    eth.h_dest[0] = 1;
    if (bpf_skb_change_head(skb, ETH_HLEN, 0)) {
        return -1;
    }
    if (bpf_skb_store_bytes(skb, 0, &eth, sizeof(eth), 0)) {
        return -1;
    }

    return bpf_redirect(skb->ifindex, BPF_F_INGRESS);
}

char _license[] SEC("license") = "GPL";
