# 拦截器

## 安装运行

编译：

```
$ make
```

拦截器可以单独运行，也可以分主从运行，主从模式下，拦截器从不用监听端口。运行方式如下：

```
$ ./b10ck agent --addr=10.67.15.8:3868
$
$ ./b10ck master --addr=10.67.15.7:3868
$ ./b10ck agent --master=10.67.15.7:3868
```

拦截器主从对外提供统一的 HTTP 接口。

```
$ curl "http://10.67.15.7:3868" -d'
{
    "domain": "$domain",
    "timeout": 300,
    "ips": ["1.2.3.4", "5.6.7.8"]
}'
```

## tcp_force_reset 程序

使用前需要先在 iptables 规则中添加规则，例如：

```
# ipset create $DOMAIN hash:ip hashsize 819200 maxelem 100000 timeout 300 -exist
# ipset add $DOMAIN $IP timeout 300 -exist;done
# iptables -A INPUT -p tcp -m tcp --dport 80 -m set --match-set $DOMAIN src -m string --string "$DOMAIN" --algo kmp --to 1480 -j NFQUEUE --queue-num 1
# iptables -I INPUT 1 -p tcp --dport 1234 -j NFQUEUE --queue-num 1
# ./tcp_force_reset 1
```

没有程序监听队列时，默认规则是丢弃所有包，当有程序监听时，就可以看到程序的工作信息：

```
# cat /proc/net/netfilter/nfnetlink_queue
1  30261     0 2 65531     0     0        0  1
```

为了避免 netlink 缓存超支（http://lists.netfilter.org/pipermail/netfilter-devel/2004-July/016182.html）
可以适当调大内核 netlink 缓存，如：

```sh
echo 524280 > /proc/sys/net/core/rmem_default
echo 524280 > /proc/sys/net/core/rmem_max
echo 524280 > /proc/sys/net/core/wmem_default
echo 524280 > /proc/sys/net/core/wmem_max
```

或编辑 /etc/sysctl.conf，加入：

```
net.core.rmem_default = 524280
net.core.rmem_max = 524280
net.core.wmem_default = 524280
net.core.wmem_max = 524280
```

目前程序实现的处理策略为：

1. tcp push、fin、ack 包设置 rst 标记，清除其他标记，截掉数据，重新计算 ip 包及 tcp 包的 checksum 后放回系统协议栈。
2. rst 包放行
3. 其他包（syn 等）丢弃

用 nc+telnet 测试效果为：

1. 到端口的新连接不能建立（客户端 SYN_SENT 至 timeout)
2. 服务端已建立的连接在收到数据时立即关闭端口（从 netstat -nt 中消失）
3. 客户端连接依旧是 ESTABLISHED，发送数据积压在 Send-Q。
