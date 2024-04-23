#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/skbuff.h>
#include <linux/types.h>
#include <linux/bpf.h>
#include <net/lwtunnel.h>
#include <net/gre.h>
#include <net/ip6_route.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/kprobes.h>
#include <net/netlink.h>
#include <net/sch_generic.h>
#include <linux/version.h>

#if LINUX_VERSION_CODE != KERNEL_VERSION(4,19,113)
#error This kernel module is only compatible/tested with 4.19.113
#endif

enum {
	BPF_CSUM_LEVEL_QUERY,
	BPF_CSUM_LEVEL_INC,
	BPF_CSUM_LEVEL_DEC,
	BPF_CSUM_LEVEL_RESET,
};

static inline void __skb_reset_checksum_unnecessary(struct sk_buff *skb)
{
	if (skb->ip_summed == CHECKSUM_UNNECESSARY) {
		skb->ip_summed = CHECKSUM_NONE;
		skb->csum_level = 0;
	}
}

// 4.19.113 使用 bpf_csum_level 函数需要用以下宏定义
// #define bpf_csum_level ((long (*)(struct __sk_buff *skb, __u64 level))73)

BPF_CALL_2(bpf_csum_level, struct sk_buff *, skb, u64, level)
{
	/* The interface is to be used in combination with bpf_skb_adjust_room()
	 * for encap/decap of packet headers when BPF_F_ADJ_ROOM_NO_CSUM_RESET
	 * is passed as flags, for example.
	 */
	printk(KERN_INFO "csum_level\n");
	switch (level) {
	case BPF_CSUM_LEVEL_INC:
		__skb_incr_checksum_unnecessary(skb);
		break;
	case BPF_CSUM_LEVEL_DEC:
		__skb_decr_checksum_unnecessary(skb);
		break;
	case BPF_CSUM_LEVEL_RESET:
		__skb_reset_checksum_unnecessary(skb);
		break;
	case BPF_CSUM_LEVEL_QUERY:
		return skb->ip_summed == CHECKSUM_UNNECESSARY ?
		       skb->csum_level : -EACCES;
	default:
		return -EINVAL;
	}

	return 0;
}

static const struct bpf_func_proto my_bpf_csum_level_proto = {
	.func		= bpf_csum_level,
	.gpl_only	= false,
	.ret_type	= RET_INTEGER,
	.arg1_type	= ARG_PTR_TO_CTX,
	.arg2_type	= ARG_ANYTHING,
};

enum my_bpf_adj_room_mode {
	//BPF_ADJ_ROOM_NET,
	BPF_ADJ_ROOM_MAC=1,
};

enum {
	BPF_F_ADJ_ROOM_FIXED_GSO	= (1ULL << 0),
	BPF_F_ADJ_ROOM_ENCAP_L3_IPV4	= (1ULL << 1),
	BPF_F_ADJ_ROOM_ENCAP_L3_IPV6	= (1ULL << 2),
	BPF_F_ADJ_ROOM_ENCAP_L4_GRE	= (1ULL << 3),
	BPF_F_ADJ_ROOM_ENCAP_L4_UDP	= (1ULL << 4),
	BPF_F_ADJ_ROOM_NO_CSUM_RESET	= (1ULL << 5),
};

enum {
	BPF_ADJ_ROOM_ENCAP_L2_MASK	= 0xff,
	BPF_ADJ_ROOM_ENCAP_L2_SHIFT	= 56,
};


static int (*my_bpf_skb_net_hdr_pop)(struct sk_buff *skb, u32 off, u32 len);
static int (*my_bpf_skb_net_hdr_push)(struct sk_buff *skb, u32 off, u32 len);

static u32 bpf_skb_net_base_len(const struct sk_buff *skb)
{
	switch (skb->protocol) {
	case htons(ETH_P_IP):
		return sizeof(struct iphdr);
	case htons(ETH_P_IPV6):
		return sizeof(struct ipv6hdr);
	default:
		return ~0U;
	}
}

static u32 __bpf_skb_max_len(const struct sk_buff *skb)
{
	return skb->dev ? skb->dev->mtu + skb->dev->hard_header_len :
			  SKB_MAX_ALLOC;
}

#define BPF_F_ADJ_ROOM_ENCAP_L2(len)	(((__u64)len & \
					  BPF_ADJ_ROOM_ENCAP_L2_MASK) \
					 << BPF_ADJ_ROOM_ENCAP_L2_SHIFT)

#define BPF_F_ADJ_ROOM_ENCAP_L3_MASK	(BPF_F_ADJ_ROOM_ENCAP_L3_IPV4 | \
					 BPF_F_ADJ_ROOM_ENCAP_L3_IPV6)


#define BPF_F_ADJ_ROOM_MASK		(BPF_F_ADJ_ROOM_FIXED_GSO | \
					 BPF_F_ADJ_ROOM_ENCAP_L3_MASK | \
					 BPF_F_ADJ_ROOM_ENCAP_L4_GRE | \
					 BPF_F_ADJ_ROOM_ENCAP_L4_UDP | \
					 BPF_F_ADJ_ROOM_ENCAP_L2( \
					  BPF_ADJ_ROOM_ENCAP_L2_MASK))

static int bpf_skb_net_shrink(struct sk_buff *skb, u32 off, u32 len_diff,
			      u64 flags)
{
	int ret;

	if (unlikely(flags & ~(BPF_F_ADJ_ROOM_FIXED_GSO |
			       BPF_F_ADJ_ROOM_NO_CSUM_RESET)))
		return -EINVAL;

	if (skb_is_gso(skb) && !skb_is_gso_tcp(skb)) {
		/* udp gso_size delineates datagrams, only allow if fixed */
		if (!(skb_shinfo(skb)->gso_type & SKB_GSO_UDP_L4) ||
		    !(flags & BPF_F_ADJ_ROOM_FIXED_GSO))
			return -ENOTSUPP;
	}

	ret = skb_unclone(skb, GFP_ATOMIC);
	if (unlikely(ret < 0))
		return ret;

	ret = my_bpf_skb_net_hdr_pop(skb, off, len_diff);
	if (unlikely(ret < 0))
		return ret;

	if (skb_is_gso(skb)) {
		struct skb_shared_info *shinfo = skb_shinfo(skb);

		/* Due to header shrink, MSS can be upgraded. */
		if (!(flags & BPF_F_ADJ_ROOM_FIXED_GSO))
			skb_increase_gso_size(shinfo, len_diff);

		/* Header must be checked, and gso_segs recomputed. */
		shinfo->gso_type |= SKB_GSO_DODGY;
		shinfo->gso_segs = 0;
	}

	return 0;
}

static int bpf_skb_net_grow(struct sk_buff *skb, u32 off, u32 len_diff,
			    u64 flags)
{
	u8 inner_mac_len = flags >> BPF_ADJ_ROOM_ENCAP_L2_SHIFT;
	bool encap = flags & BPF_F_ADJ_ROOM_ENCAP_L3_MASK;
	u16 mac_len = 0, inner_net = 0, inner_trans = 0;
	unsigned int gso_type = SKB_GSO_DODGY;
	int ret;

	if (skb_is_gso(skb) && !skb_is_gso_tcp(skb)) {
		/* udp gso_size delineates datagrams, only allow if fixed */
		if (!(skb_shinfo(skb)->gso_type & SKB_GSO_UDP_L4) ||
		    !(flags & BPF_F_ADJ_ROOM_FIXED_GSO))
			return -ENOTSUPP;
	}

	ret = skb_cow_head(skb, len_diff);
	if (unlikely(ret < 0))
		return ret;

	if (encap) {
		if (skb->protocol != htons(ETH_P_IP) &&
		    skb->protocol != htons(ETH_P_IPV6))
			return -ENOTSUPP;

		if (flags & BPF_F_ADJ_ROOM_ENCAP_L3_IPV4 &&
		    flags & BPF_F_ADJ_ROOM_ENCAP_L3_IPV6)
			return -EINVAL;

		if (flags & BPF_F_ADJ_ROOM_ENCAP_L4_GRE &&
		    flags & BPF_F_ADJ_ROOM_ENCAP_L4_UDP)
			return -EINVAL;

		if (skb->encapsulation)
			return -EALREADY;

		mac_len = skb->network_header - skb->mac_header;
		inner_net = skb->network_header;
		if (inner_mac_len > len_diff)
			return -EINVAL;
		inner_trans = skb->transport_header;
	}

	ret = my_bpf_skb_net_hdr_push(skb, off, len_diff);
	if (unlikely(ret < 0))
		return ret;

	if (encap) {
		skb->inner_mac_header = inner_net - inner_mac_len;
		skb->inner_network_header = inner_net;
		skb->inner_transport_header = inner_trans;
		skb_set_inner_protocol(skb, skb->protocol);

		skb->encapsulation = 1;
		skb_set_network_header(skb, mac_len);

		if (flags & BPF_F_ADJ_ROOM_ENCAP_L4_UDP)
			gso_type |= SKB_GSO_UDP_TUNNEL;
		else if (flags & BPF_F_ADJ_ROOM_ENCAP_L4_GRE)
			gso_type |= SKB_GSO_GRE;
		else if (flags & BPF_F_ADJ_ROOM_ENCAP_L3_IPV6)
			gso_type |= SKB_GSO_IPXIP6;
		else if (flags & BPF_F_ADJ_ROOM_ENCAP_L3_IPV4)
			gso_type |= SKB_GSO_IPXIP4;

		if (flags & BPF_F_ADJ_ROOM_ENCAP_L4_GRE ||
		    flags & BPF_F_ADJ_ROOM_ENCAP_L4_UDP) {
			int nh_len = flags & BPF_F_ADJ_ROOM_ENCAP_L3_IPV6 ?
					sizeof(struct ipv6hdr) :
					sizeof(struct iphdr);

			skb_set_transport_header(skb, mac_len + nh_len);
		}

		/* Match skb->protocol to new outer l3 protocol */
		if (skb->protocol == htons(ETH_P_IP) &&
		    flags & BPF_F_ADJ_ROOM_ENCAP_L3_IPV6)
			skb->protocol = htons(ETH_P_IPV6);
		else if (skb->protocol == htons(ETH_P_IPV6) &&
			 flags & BPF_F_ADJ_ROOM_ENCAP_L3_IPV4)
			skb->protocol = htons(ETH_P_IP);
	}

	if (skb_is_gso(skb)) {
		struct skb_shared_info *shinfo = skb_shinfo(skb);

		/* Due to header grow, MSS needs to be downgraded. */
		if (!(flags & BPF_F_ADJ_ROOM_FIXED_GSO))
			skb_decrease_gso_size(shinfo, len_diff);

		/* Header must be checked, and gso_segs recomputed. */
		shinfo->gso_type |= gso_type;
		shinfo->gso_segs = 0;
	}

	return 0;
}

BPF_CALL_4(my_bpf_skb_adjust_room, struct sk_buff *, skb, s32, len_diff,
	   u32, mode, u64, flags)
{
	printk(KERN_INFO "bpf_skb_adjust_room(%p, %d, %d, %x)\n", skb, len_diff, mode, flags);
	u32 len_cur, len_diff_abs = abs(len_diff);
	u32 len_min = bpf_skb_net_base_len(skb);
	u32 len_max = __bpf_skb_max_len(skb);
	__be16 proto = skb->protocol;
	bool shrink = len_diff < 0;
	u32 off;
	int ret;

	if (unlikely(flags & ~(BPF_F_ADJ_ROOM_MASK |
			       BPF_F_ADJ_ROOM_NO_CSUM_RESET)))
		return -EINVAL;
	if (unlikely(len_diff_abs > 0xfffU))
		return -EFAULT;
	if (unlikely(proto != htons(ETH_P_IP) &&
		     proto != htons(ETH_P_IPV6)))
		return -ENOTSUPP;

	off = skb_mac_header_len(skb);
	switch (mode) {
	case BPF_ADJ_ROOM_NET:
		off += bpf_skb_net_base_len(skb);
		break;
	case BPF_ADJ_ROOM_MAC:
		break;
	default:
		return -ENOTSUPP;
	}

	len_cur = skb->len - skb_network_offset(skb);
	if ((shrink && (len_diff_abs >= len_cur ||
			len_cur - len_diff_abs < len_min)) ||
	    (!shrink && (skb->len + len_diff_abs > len_max &&
			 !skb_is_gso(skb))))
		return -ENOTSUPP;

	ret = shrink ? bpf_skb_net_shrink(skb, off, len_diff_abs, flags) :
		       bpf_skb_net_grow(skb, off, len_diff_abs, flags);
	if (!ret && !(flags & BPF_F_ADJ_ROOM_NO_CSUM_RESET))
		__skb_reset_checksum_unnecessary(skb);

	bpf_compute_data_pointers(skb);
	return ret;
}

static const struct bpf_func_proto my_bpf_skb_adjust_room_proto = {
	.func		= my_bpf_skb_adjust_room,
	.gpl_only	= false,
	.ret_type	= RET_INTEGER,
	.arg1_type	= ARG_PTR_TO_CTX,
	.arg2_type	= ARG_ANYTHING,
	.arg3_type	= ARG_ANYTHING,
	.arg4_type	= ARG_ANYTHING,
};

static int handler_post(struct kretprobe_instance *ri, struct pt_regs *regs)
{
	printk(KERN_INFO "kretprobe tc_cls_act_func_proto func_id:%d, retval: %p\n", (int)regs->di, (void*)regs->ax);
	if (regs->di == 73 /*BPF_FUNC_csum_level*/) {
		regs->ax = (unsigned long)&my_bpf_csum_level_proto;
	}
	else if (regs->di == BPF_FUNC_skb_adjust_room) {
		printk(KERN_INFO "hook skb_adjust_room:%d, retval: %p\n", (int)regs->di, (void*)regs->ax);
		regs->ax = (unsigned long)&my_bpf_skb_adjust_room_proto;
	}
	return 0;
}

static struct kretprobe my_kretprobe = {
	.handler    = handler_post,
};

int init_module(void)
{

	my_bpf_skb_net_hdr_pop = kallsyms_lookup_name("bpf_skb_net_hdr_pop");
	my_bpf_skb_net_hdr_push = kallsyms_lookup_name("bpf_skb_net_hdr_push");
	if (my_bpf_skb_net_hdr_pop == NULL || my_bpf_skb_net_hdr_push == NULL) {
		printk(KERN_INFO "NULL");
		return 1;
	}
	my_kretprobe.kp.addr = (kprobe_opcode_t *)kallsyms_lookup_name("tc_cls_act_func_proto");
	register_kretprobe(&my_kretprobe);

	return 0;
}

void cleanup_module(void)
{
	unregister_kretprobe(&my_kretprobe);
}

MODULE_LICENSE("GPL");

