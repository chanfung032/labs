package t

import (
	"bytes"
	"encoding/binary"
	"net"
	"syscall"
	"testing"
)

func TestSendWithIPHdrIncl(t *testing.T) {
	srcPort := uint16(12345)
	dstPort := uint16(80)
	seqNumber := uint32(1000)

	srcIP := net.ParseIP("192.168.1.2")
	dstIP := net.ParseIP("9.9.9.9")
	if err := SendRSTPacket(srcIP, dstIP, srcPort, dstPort, seqNumber); err != nil {
		t.Fatal("ipv4:", err)
	}

	srcIP = net.ParseIP("2410:8c00:6c21:1051:0:ff:b0af:279a")
	dstIP = net.ParseIP("2411:8c00:6c21:1051:0:ff:b0af:279a")
	if err := SendRSTPacket(srcIP, dstIP, srcPort, dstPort, seqNumber); err != nil {
		t.Fatal("ipv6:", err)
	}
}

func SendRSTPacket(srcIP, dstIP net.IP, srcPort, dstPort uint16, seqNumber uint32) error {
	var pkt []byte
	var tcp []byte
	var family int
	if srcIP.To4() != nil {
		pkt = []byte{
			0x45, 0x00, 0x00, 0x28, 0x00, 0x00, 0x40, 0x00,
			0xff, 0x06, 0x00, 0x00, 0xc0, 0xa8, 0x01, 0x02,
			0x08, 0x08, 0x08, 0x08, 0x30, 0x39, 0x00, 0x50,
			0x00, 0x00, 0x03, 0xe8, 0x00, 0x00, 0x00, 0x00,
			0x50, 0x04, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00,
		}
		copy(pkt[12:16], srcIP.To4())
		copy(pkt[16:20], dstIP.To4())
		binary.BigEndian.PutUint16(pkt[10:12], calculateChecksum(pkt[:20]))
		tcp = pkt[20:]
		family = syscall.AF_INET
	} else {
		pkt = []byte{
			0x60, 0x00, 0x00, 0x00, 0x00, 0x14, 0x06, 0xff,
			0x24, 0x10, 0x8c, 0x00, 0x6c, 0x21, 0x10, 0x51,
			0x00, 0x00, 0x00, 0xff, 0xb0, 0xaf, 0x27, 0x9a,
			0x24, 0x09, 0x8c, 0x00, 0x6c, 0x21, 0x10, 0x51,
			0x00, 0x00, 0x00, 0xff, 0xb0, 0xaf, 0x27, 0x9a,
			0x30, 0x39, 0x00, 0x50, 0x00, 0x00, 0x03, 0xe8,
			0x00, 0x00, 0x00, 0x00, 0x50, 0x04, 0x02, 0x00,
			0x00, 0x00, 0x00, 0x00,
		}
		copy(pkt[8:24], srcIP.To16())
		copy(pkt[24:40], dstIP.To16())
		tcp = pkt[40:]
		family = syscall.AF_INET6
	}
	binary.BigEndian.PutUint16(tcp[0:2], srcPort)
	binary.BigEndian.PutUint16(tcp[2:4], dstPort)
	binary.BigEndian.PutUint32(tcp[4:8], seqNumber)
	pseudoHeader := createPseudoHeader(srcIP, dstIP, tcp)
	checksum := calculateChecksum(pseudoHeader)
	binary.BigEndian.PutUint16(tcp[16:18], checksum)

	if family == syscall.AF_INET {
		conn, err := syscall.Socket(family, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
		if err != nil {
			return err
		}
		err = syscall.SetsockoptInt(conn, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
		if err != nil {
			return err
		}
		addr := &syscall.SockaddrInet4{Port: int(dstPort)}
		copy(addr.Addr[:], dstIP.To4())
		return syscall.Sendto(conn, pkt, 0, addr)
	} else {
		conn, err := syscall.Socket(family, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
		if err != nil {
			return err
		}
		err = syscall.SetsockoptInt(conn, syscall.IPPROTO_IPV6, 36 /*IPV6_HDRINCL*/, 1)
		if err != nil {
			return err
		}
		addr := &syscall.SockaddrInet6{}
		copy(addr.Addr[:], dstIP)
		return syscall.Sendto(conn, pkt, 0, addr)
	}
}

func createPseudoHeader(srcIP, dstIP net.IP, tcpHeader []byte) []byte {
	pseudo := &bytes.Buffer{}
	if srcIP.To4() != nil {
		pseudo.Write(srcIP.To4()) // Source IP
		pseudo.Write(dstIP.To4()) // Destination IP
		pseudo.WriteByte(0)       // Reserved
		pseudo.WriteByte(syscall.IPPROTO_TCP)
		binary.Write(pseudo, binary.BigEndian, uint16(len(tcpHeader)))
	} else {
		pseudo.Write(srcIP.To16()) // Source IP
		pseudo.Write(dstIP.To16()) // Destination IP
		binary.Write(pseudo, binary.BigEndian, uint32(len(tcpHeader)))
		pseudo.WriteByte(0) // Reserved
		pseudo.WriteByte(syscall.IPPROTO_TCP)
	}

	pseudo.Write(tcpHeader)
	return pseudo.Bytes()
}

func calculateChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	return ^uint16(sum)
}
