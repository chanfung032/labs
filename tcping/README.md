# Tcping

Raw socket Tcping 实现。修改自：https://github.com/grahamking/latency, 修改封包
解包使用 gopacket 库，添加了 raw socket 包过滤解决大流量机器上收到的包太多 CPU
高的问题。
