package tls

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

func TestHTTPS(t *testing.T) {
	c, err := net.Dial("tcp", "www.cloudflare.com:443")
	if err != nil {
		t.Fatal("connect failed", err)
	}
	tlsConn := &TLS13Conn{Conn: c}
	tlsConn.Handshake()
	tlsConn.Write([]byte("GET / HTTP/1.1\r\nHost: www.cloudflare.com\r\n\r\n"))
	tlsConn.Read()
	resp := tlsConn.Read()
	fmt.Printf("%s\n", resp)
	if !bytes.HasPrefix(resp, []byte("HTTP/1.1 200 OK")) {
		t.FailNow()
	}
}
