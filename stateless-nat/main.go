package main

/*
编译方法：

go generate
#编译调试版本
#BPF_CFLAGS="-DDEBUG" go generate
go build

运行方法见 `-h` 帮助信息。

---

所以需求就是实现一个无状态版本的dnat是吧？

也就是假设有规则：
1.2.3.4:80 -> 192.168.1.4:80

那么：

进来的包的目的IP端口是1.2.3.4:80，修改其目的IP端口为 192.168.1.4:80 源IP为本机IP
发出的包的源IP端口是192.168.1.4:80，修改源IP端口为1.2.3.4:80

---

测试环境测试了下 似乎没问题了

我的环境是这样的：

71上 em1 网卡绑了两个IP 172.18.22.71 和 20.20.20.6，启动一个 web server： python3 -m http.server --bind 20.20.20.6 80

设置dnat： ./stateless_nat -i em1 172.18.22.71:8080-20.20.20.6:80

然后在72机器上 curl 172.18.22.71:8080，可以收到上面 python web server的返回。
*/

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cflags "$BPF_CFLAGS" nat nat.c -- -O2

import (
	"encoding/binary"
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

/*
#include "nat.h"
*/
import "C"

func ip2int(s string) uint32 {
	ip := net.ParseIP(s)
	if ip == nil {
		return 0
	}
	var ipv4 uint32
	if ip4 := ip.To4(); ip4 != nil {
		ipv4 = binary.LittleEndian.Uint32(ip4)
		return ipv4
	} else {
		return 0
	}
}

func port2int(s string) uint16 {
	i, _ := strconv.Atoi(s)
	return uint16(i)
}

func checkOrSetupQdisc(link netlink.Link) error {
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return fmt.Errorf("Listing qdisc failed: %v", err)
	}
	for _, q := range qdiscs {
		if q.Type() == "clsact" {
			log.Debug("qdisc clsact found")
			return nil
		}
	}

	attrs := netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(0xffff, 0),
		Parent:    netlink.HANDLE_CLSACT,
	}

	qdisc := &netlink.GenericQdisc{
		QdiscAttrs: attrs,
		QdiscType:  "clsact",
	}

	if err = netlink.QdiscReplace(qdisc); err != nil {
		return fmt.Errorf("Replacing qdisc failed: %v", err)
	}

	return nil
}

func main() {
	iface := flag.String("i", "", "Network interface(s) to hook xdp program")
	flag.Parse()

	if len(*iface) == 0 {
		fmt.Printf(`%s -i "IFACE1 IFACE2 ..." IP:PORT-IP:PORT ...

`, os.Args[0])
		return
	}

	ifaces := strings.Fields(*iface)
	log.Info("iface: ", ifaces)

	if len(flag.Args()) == 0 {
		// 如果没有 nat 规则，那么卸载 bpf 程序
		for _, iface := range ifaces {
			link, _ := netlink.LinkByName(iface)
			filterattrs := netlink.FilterAttrs{
				LinkIndex: link.Attrs().Index,
				Parent:    netlink.HANDLE_MIN_INGRESS,
				Handle:    netlink.MakeHandle(0, 1),
				Protocol:  unix.ETH_P_ALL,
				Priority:  1,
			}
			filter := netlink.BpfFilter{
				FilterAttrs: filterattrs,
			}
			if err := netlink.FilterDel(&filter); err != nil {
				log.Warnf("tc bpf filter del %s ingress failed: %v", iface, err)
			} else {
				log.Infof("tc bpf filter del %s ingress ... ok", iface)
			}

			filterattrs = netlink.FilterAttrs{
				LinkIndex: link.Attrs().Index,
				Parent:    netlink.HANDLE_MIN_EGRESS,
				Handle:    netlink.MakeHandle(0, 1),
				Protocol:  unix.ETH_P_ALL,
				Priority:  1,
			}
			filter = netlink.BpfFilter{
				FilterAttrs: filterattrs,
			}
			if err := netlink.FilterDel(&filter); err != nil {
				log.Warnf("tc bpf filter del %s egress failed: %v", iface, err)
			} else {
				log.Infof("tc bpf filter del %s egress ... ok", iface)
			}
		}
		return
	}

	// 解析 nat 规则
	bindRe := regexp.MustCompile(`^(.*):(.*)-(.*):(.*)$`)
	mapping := make([][2]C.bind_t, len(flag.Args()))
	for i, bind := range flag.Args() {
		m := bindRe.FindStringSubmatch(bind)
		if m == nil {
			log.Fatal("invalid args: ", bind)
		}

		sip := ip2int(m[1])
		sport := port2int(m[2])
		dip := ip2int(m[3])
		dport := port2int(m[4])
		if sip == 0 || sport == 0 || dip == 0 || dport == 0 {
			log.Fatal("invalid ip:port: ", bind)
		}

		mapping[i][0] = C.bind_t{
			ipv4: C.uint(sip),
			port: C.ushort(sport<<8 + sport>>8),
		}
		mapping[i][1] = C.bind_t{
			ipv4: C.uint(dip),
			port: C.ushort(dport<<8 + dport>>8),
		}
	}

	if err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	}); err != nil {
		log.Fatal(err)
	}

	// 加载 ebpf 程序
	var objs natObjects
	if err := loadNatObjects(&objs, nil); err != nil {
	}
	defer objs.Close()
	log.Info(objs)

	// 存储 nat 规则到 bpf map 中
	dTable := objs.natMaps.DnatMapping
	sTable := objs.natMaps.SnatMapping
	for _, m := range mapping {
		if err := dTable.Put(unsafe.Pointer(&m[0]), unsafe.Pointer(&m[1])); err != nil {
			log.Fatal("bTable put failed: ", err)
		}
		if err := sTable.Put(unsafe.Pointer(&m[1]), unsafe.Pointer(&m[0])); err != nil {
			log.Fatal("bTable put failed: ", err)
		}
	}

	ingressFd := objs.natPrograms.BpfIngress.FD()
	egressFd := objs.natPrograms.BpfEgress.FD()

	for _, iface := range ifaces {
		log.Info("inject: ", iface)
		link, err := netlink.LinkByName(iface)
		if err != nil {
			log.Fatal(err)
		}

		if err := checkOrSetupQdisc(link); err != nil {
			log.Fatal(err)
		}

		// 挂载 bpf 程序到 tc ingress hook 点
		filterattrs := netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.HANDLE_MIN_INGRESS,
			Handle:    netlink.MakeHandle(0, 1),
			Protocol:  unix.ETH_P_ALL,
			Priority:  1,
		}
		filter := netlink.BpfFilter{
			FilterAttrs:  filterattrs,
			Fd:           ingressFd,
			Name:         "bpf-dnat",
			DirectAction: true,
		}
		if err := netlink.FilterReplace(&filter); err != nil {
			log.Fatal("tc bpf filter create or replace failed: ", err)
		}

		// 挂载 bpf 程序到 tc egress hook 点
		filterattrs = netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.HANDLE_MIN_EGRESS,
			Handle:    netlink.MakeHandle(0, 1),
			Protocol:  unix.ETH_P_ALL,
			Priority:  1,
		}
		filter = netlink.BpfFilter{
			FilterAttrs:  filterattrs,
			Fd:           egressFd,
			Name:         "bpf-snat",
			DirectAction: true,
		}
		if err := netlink.FilterReplace(&filter); err != nil {
			log.Fatal("tc bpf filter create or replace failed: ", err)
		}
	}
}
