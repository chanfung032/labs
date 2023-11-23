#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/kprobes.h>
#include <linux/fs.h>
#include <linux/kallsyms.h>

static struct hlist_bl_head *dentry_hashtable;
static unsigned int d_hash_shift;

// linux-4.14.69，不同版本的 d_hash 可能会不一样
// https://elixir.bootlin.com/linux/v4.14.69/source/fs/dcache.c#L112
static inline struct hlist_bl_head *d_hash(unsigned int hash)
{
    return dentry_hashtable + (hash >> (32 - d_hash_shift));
}

static inline long hlist_count(const struct dentry *parent, const struct qstr *name)
{
    long count = 0;

    // 和 __d_lookup 函数中对应的查询循环逻辑类似
    // https://elixir.bootlin.com/linux/v4.14.69/source/fs/dcache.c#L2281
    unsigned int hash = name->hash;
    struct hlist_bl_head *b = d_hash(hash);
    struct hlist_bl_node *node;
    struct dentry *dentry;
    rcu_read_lock();
    hlist_bl_for_each_entry_rcu(dentry, node, b, d_hash) {
        count++;
    }
    rcu_read_unlock();

    if (count > 2) {
        printk("hlist_bl_head=%p, count=%ld, name=%s, hash=%u\n",b, count, name->name, name->hash);
    }

    return count;
}

static int __kprobes handler(struct kprobe *p, struct pt_regs *regs)
{
    int count = hlist_count(regs->di, regs->si);
    return 0;
}

struct kprobe kp = {
    .symbol_name = "__d_lookup",
    .pre_handler = handler,
};

static int __init kprobe_init(void)
{
    void *addr;
    addr = kallsyms_lookup_name("dentry_hashtable");
    if (!addr) {
        pr_err("unresolved dentry_hashtable");
        return -EINVAL;
    }
    dentry_hashtable = *(struct hlist_bl_head **)addr;

    addr = kallsyms_lookup_name("d_hash_shift");
    if (!addr) {
        pr_err("unresolved d_hash_shift");
        return -EINVAL;
    }
    d_hash_shift = *(unsigned int*)addr;

    int ret;
    ret = register_kprobe(&kp);
    if (ret < 0) {
        pr_err("register_kprobe failed, returned %d\n", ret);
        return ret;
    }
    pr_info("Planted kprobe at %p\n", kp.addr);

    return 0;
}

static void __exit kprobe_exit(void)
{
    unregister_kprobe(&kp);
    pr_info("kprobe at %p unregistered\n", kp.addr);
}

module_init(kprobe_init)
module_exit(kprobe_exit)
MODULE_LICENSE("GPL");
