package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/sys/unix"
)

var (
	ifaceParam   = flag.String("i", "", "Interface (e.g. eth0, wlan1, etc)")
	portParam    = flag.Int("p", 80, "Port to test against (default 80)")
	lportParam   = flag.Int("lp", 0xaa47, "Local port to bind for test(default 0xaa47")
	autoParam    = flag.Bool("a", false, "Tcping to some well known addresses")
	defaultHosts = map[string]string{
		"Baidu":  "www.baidu.com",
		"Taobao": "www.taobao.com",
		"AWS":    "aws.amazon.com",
	}
)

func main() {
	flag.Usage = usage
	flag.Parse()

	iface := *ifaceParam
	if iface == "" {
		iface = chooseInterface()
		if iface == "" {
			fmt.Println("Could not decide which net interface to use.")
			fmt.Println("Specify it with -i <iface> param")
			os.Exit(1)
		}
	}

	localAddr := interfaceAddress(iface)
	laddr := strings.Split(localAddr.String(), "/")[0]

	port := uint16(*portParam)
	if *autoParam {
		autoTest(laddr, port)
		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Missing remote address")
		usage()
		os.Exit(1)
	}

	remoteHost := flag.Arg(0)
	fmt.Println("Tcping from", laddr, "to", remoteHost, "on port", port)
	fmt.Printf("Latency: %v\n", tcping(laddr, remoteHost, port))
}

func autoTest(localAddr string, port uint16) {
	for name, host := range defaultHosts {
		fmt.Printf("%15s: %v\n", name, tcping(localAddr, host, port))
	}
}

func tcping(localAddr string, remoteHost string, port uint16) time.Duration {
	var wg sync.WaitGroup
	wg.Add(1)
	var receiveTime time.Time

	addrs, err := net.LookupHost(remoteHost)
	if err != nil {
		log.Fatalf("Error resolving %s. %s\n", remoteHost, err)
	}
	remoteAddr := addrs[0]

	go func() {
		receiveTime = receiveSynAck(localAddr, remoteAddr)
		wg.Done()
	}()

	time.Sleep(1 * time.Millisecond)
	sendTime := sendSyn(localAddr, remoteAddr, port)

	wg.Wait()
	return receiveTime.Sub(sendTime)
}

func chooseInterface() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("net.Interfaces: %s", err)
	}
	for _, iface := range interfaces {
		// Skip loopback
		if iface.Name == "lo" {
			continue
		}
		addrs, err := iface.Addrs()
		// Skip if error getting addresses
		if err != nil {
			log.Println("Error get addresses for interfaces %s. %s", iface.Name, err)
			continue
		}

		if len(addrs) > 0 {
			// This one will do
			return iface.Name
		}
	}

	return ""
}

func interfaceAddress(ifaceName string) net.Addr {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("net.InterfaceByName for %s. %s", ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Fatalf("iface.Addrs: %s", err)
	}
	return addrs[0]
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <remote>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func sendSyn(laddr, raddr string, port uint16) time.Time {
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(*lportParam),
		DstPort: layers.TCPPort(port),
		SYN:     true,          // Set SYN flag
		Window:  14600,         // Common TCP window size
		Seq:     rand.Uint32(), // Random sequence number
	}
	tcp.SetNetworkLayerForChecksum(&layers.IPv4{
		SrcIP:    net.ParseIP(laddr),
		DstIP:    net.ParseIP(raddr),
		Protocol: layers.IPProtocolTCP,
	})
	buffer := gopacket.NewSerializeBuffer()
	err := tcp.SerializeTo(buffer, gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true})
	if err != nil {
		log.Fatal("tcp seralization", err)
	}
	log.Printf("-> %v seq=%d\n", raddr, tcp.Seq)
	data := buffer.Bytes()

	/*
		// Bind local port before send syn
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
		if err != nil {
			panic(fmt.Sprintf("Failed to create socket: %v", err))
		}
		defer syscall.Close(fd)
		addr := &syscall.SockaddrInet4{Port: *lportParam}
		copy(addr.Addr[:], net.ParseIP(laddr).To4()) // Specify the local IP
		if err := syscall.Bind(fd, addr); err != nil {
			panic(fmt.Sprintf("Failed to bind to port %d: %v", port, err))
		}
	*/

	conn, err := net.Dial("ip4:tcp", raddr)
	if err != nil {
		log.Fatalf("Dial: %s\n", err)
	}

	sendTime := time.Now()

	numWrote, err := conn.Write(data)
	if err != nil {
		log.Fatalf("Write: %s\n", err)
	}
	if numWrote != len(data) {
		log.Fatalf("Short write. Wrote %d/%d bytes\n", numWrote, len(data))
	}

	conn.Close()

	return sendTime
}

// BPF program to filter TCP packets to local bind port
func installBpfFilter(fd int) {
	var bpfProgram = []syscall.SockFilter{
		// load tcp dst port
		{Code: syscall.BPF_LD + syscall.BPF_H + syscall.BPF_ABS, K: 22},
		// compare & jmp
		{Code: syscall.BPF_JMP + syscall.BPF_JEQ + syscall.BPF_K, K: uint32(*lportParam), Jt: 0, Jf: 1},
		// accept packet
		{Code: syscall.BPF_RET + syscall.BPF_K, K: 0xFFFFFFFF},
		// reject packet
		{Code: syscall.BPF_RET + syscall.BPF_K, K: 0},
	}

	prog := syscall.SockFprog{
		Len:    uint16(len(bpfProgram)),
		Filter: (*syscall.SockFilter)(unsafe.Pointer(&bpfProgram[0])),
	}
	if _, _, errno := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd), uintptr(syscall.SOL_SOCKET), uintptr(syscall.SO_ATTACH_FILTER),
		uintptr(unsafe.Pointer(&prog)), uintptr(unix.SizeofSockFprog), 0); errno != 0 {
		log.Fatalf("setsockopt err:%v", errno)
	}
}

func receiveSynAck(localAddress, remoteAddress string) time.Time {
	netaddr, err := net.ResolveIPAddr("ip4", localAddress)
	if err != nil {
		log.Fatalf("net.ResolveIPAddr: %s. %s\n", localAddress, netaddr)
	}

	conn, err := net.ListenIP("ip4:tcp", netaddr)
	if err != nil {
		log.Fatalf("ListenIP: %s\n", err)
	}
	rawconn, _ := conn.SyscallConn()
	rawconn.Control(func(fd uintptr) { installBpfFilter(int(fd)) })

	var receiveTime time.Time
	for {
		buf := make([]byte, 1024)
		numRead, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatalf("ReadFrom: %s\n", err)
		}
		if raddr.String() != remoteAddress {
			// this is not the packet we are looking for
			fmt.Println("not not: ", raddr)
			continue
		}
		receiveTime = time.Now()

		//fmt.Printf("Received: %v, %x\n", raddr, buf[:numRead])
		packet := gopacket.NewPacket(buf[:numRead], layers.LayerTypeTCP, gopacket.Default)
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if tcp.RST || (tcp.SYN && tcp.ACK) {
				log.Printf("<- %v ack=%d\n", remoteAddress, tcp.Ack)
				break
			}
		}
	}
	return receiveTime
}
