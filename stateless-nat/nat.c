#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <linux/filter.h>
#include <linux/pkt_cls.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <bpf/bpf_helpers.h>
#include "nat.h"

#ifdef DEBUG
#define dd(...) bpf_printk(__VA_ARGS__)
#else
#define dd(...)
#endif

#define offsetof(TYPE, MEMBER)	((size_t)&((TYPE *)0)->MEMBER)

// unsigned long long load_byte(void *skb,
// 			     unsigned long long off) asm("llvm.bpf.load.byte");
// unsigned long long load_half(void *skb,
// 			     unsigned long long off) asm("llvm.bpf.load.half");
// unsigned long long load_word(void *skb,
// 			     unsigned long long off) asm("llvm.bpf.load.word");

struct bpf_map_def SEC("maps") dnat_mapping = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(bind_t),
    .value_size = sizeof(bind_t),
    .max_entries = 4096,
};

struct bpf_map_def SEC("maps") snat_mapping = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(bind_t),
    .value_size = sizeof(bind_t),
    .max_entries = 4096,
};

#define _htonl __builtin_bswap32
#define _htons __builtin_bswap16

#define IP_CSUM_OFF (ETH_HLEN + offsetof(struct iphdr, check))

#define TCP_CSUM_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, check))
#define IP_SRC_OFF (ETH_HLEN + offsetof(struct iphdr, saddr))

#define IS_PSEUDO 0x10

#define IP_SRC_OFF (ETH_HLEN + offsetof(struct iphdr, saddr))
static inline void set_tcp_src_ip(struct __sk_buff *skb, __u32 old_ip, __u32 new_ip)
{
	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_ip, new_ip, IS_PSEUDO | sizeof(new_ip));
	bpf_l3_csum_replace(skb, IP_CSUM_OFF, old_ip, new_ip, sizeof(new_ip));
	bpf_skb_store_bytes(skb, IP_SRC_OFF, &new_ip, sizeof(new_ip), 0);
}

#define TCP_SPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, source))
static inline void set_tcp_src_port(struct __sk_buff *skb, __u16 old_port, __u16 new_port)
{
	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_port, new_port, sizeof(new_port));
	bpf_skb_store_bytes(skb, TCP_SPORT_OFF, &new_port, sizeof(new_port), 0);
}

#define IP_DEST_OFF (ETH_HLEN + offsetof(struct iphdr, daddr))
static inline void set_tcp_dst_ip(struct __sk_buff *skb, __u32 old_ip, __u32 new_ip)
{
	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_ip, new_ip, IS_PSEUDO | sizeof(new_ip));
	bpf_l3_csum_replace(skb, IP_CSUM_OFF, old_ip, new_ip, sizeof(new_ip));
	bpf_skb_store_bytes(skb, IP_DEST_OFF, &new_ip, sizeof(new_ip), 0);
}

#define TCP_DPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, dest))
static inline void set_tcp_dst_port(struct __sk_buff *skb, __u16 old_port, __u16 new_port)
{
	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_port, new_port, sizeof(new_port));
	bpf_skb_store_bytes(skb, TCP_DPORT_OFF, &new_port, sizeof(new_port), 0);
}

#define HLEN (sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct tcphdr))

SEC("classifier/ingress")
int bpf_ingress(struct __sk_buff *skb)
{
    dd("========= ingress ");
    if (skb->protocol != _htons(ETH_P_IP)) {
        return TC_ACT_OK;
    }

    if (bpf_skb_pull_data(skb, HLEN)) {
        return TC_ACT_OK;
    }

    void *data_end = (void *)(unsigned long long)skb->data_end;
	void *data = (void *)(unsigned long long)skb->data;

    struct iphdr *ip = data + sizeof(struct ethhdr);
    if (ip + 1 > data_end)
        return TC_ACT_OK;

    struct tcphdr *tcp = ip + 1;
    if (tcp + 1 > data_end)
        return TC_ACT_OK;

	if (ip->protocol != IPPROTO_TCP) {
        dd("protocol %d", ip->protocol);
        return TC_ACT_OK;
    }

    bind_t key, value, *value_p;

    key.ipv4 = ip->daddr;
    key.port = tcp->dest;
    value_p = (bind_t*)bpf_map_lookup_elem(&dnat_mapping, &key);
    if (value_p != NULL) {
        dd("rewrite dst");
        value = *value_p;
        set_tcp_dst_ip(skb, key.ipv4, value.ipv4);
        set_tcp_dst_port(skb, key.port, value.port);
        return TC_ACT_OK;
    }

    return TC_ACT_OK;
}

SEC("classifier/egress")
int bpf_egress(struct __sk_buff *skb)
{
    dd("-------- egress ");

    if (skb->protocol != _htons(ETH_P_IP)) {
        return TC_ACT_OK;
    }

    if (bpf_skb_pull_data(skb, HLEN)) {
        return TC_ACT_OK;
    }

    void *data_end = (void *)(unsigned long long)skb->data_end;
	void *data = (void *)(unsigned long long)skb->data;

    struct iphdr *ip = data + sizeof(struct ethhdr);
    if (ip + 1 > data_end)
        return TC_ACT_OK;

    struct tcphdr *tcp = ip + 1;
    if (tcp + 1 > data_end)
        return TC_ACT_OK;

	if (ip->protocol != IPPROTO_TCP) {
        dd("protocol %d", ip->protocol);
        return TC_ACT_OK;
    }

    bind_t key, value, *value_p;

    key.ipv4 = ip->saddr;
    key.port = tcp->source;
    value_p = (bind_t*)bpf_map_lookup_elem(&snat_mapping, &key);
    if (value_p != NULL) {
        dd("rewrite src");
        value = *value_p;
        set_tcp_src_ip(skb, key.ipv4, value.ipv4);
        set_tcp_src_port(skb, key.port, value.port);
        return TC_ACT_OK;
    }

    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";
