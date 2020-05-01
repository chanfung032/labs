package tls

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"io"
	"net"
	"strings"
)

type TLS13Conn struct {
	net.Conn

	// 客户端使用以下 Key, IV 加密发送给服务端的上行数据
	clientApplicationKey []byte
	clientApplicationIV  []byte
	// 客户端使用以下 Key, IV 解密服务端发来的下行数据
	serverApplicationKey []byte
	serverApplicationIV  []byte
	// 客户端发送和接收的 Record 计数
	clientRecordNum int
	serverRecordNum int
}

func (c *TLS13Conn) Handshake() {
	// ÷ 生成一对公钥私钥
	clientPrivateKey := make([]byte, curve25519.ScalarSize)
	io.ReadFull(rand.Reader, clientPrivateKey)
	clientPublicKey, _ := curve25519.X25519(clientPrivateKey, curve25519.Basepoint)
	//clientPrivateKey := hex2byte(`202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f`)
	//clientPublicKey := hex2byte(`358072d6365880d1aeea329adf9121383851ed21a28e3b75e965d0d2cd166254`)
	fmt.Printf("clientPrivateKey: %x\nclientPublicKey: %x\n", clientPrivateKey, clientPublicKey)

	// ⇉ 发送 ClientHello
	// TODO: SNI support
	clientHello := hex2byte(`16 03 01 00 ca 01 00 00 c6 03 03 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17 18 19 1a 1b 1c 1d 1e 1f 20 e0 e1 e2 e3 e4 e5 e6 e7 e8 e9 ea eb ec ed ee ef f0 f1 f2 f3 f4 f5 f6 f7 f8 f9 fa fb fc fd fe ff 00 06 13 01 13 02 13 03 01 00 00 77 00 00 00 18 00 16 00 00 13 65 78 61 6d 70 6c 65 2e 75 6c 66 68 65 69 6d 2e 6e 65 74 00 0a 00 08 00 06 00 1d 00 17 00 18 00 0d 00 14 00 12 04 03 08 04 04 01 05 03 08 05 05 01 08 06 06 01 02 01 00 33 00 26 00 24 00 1d 00 20 35 80 72 d6 36 58 80 d1 ae ea 32 9a df 91 21 38 38 51 ed 21 a2 8e 3b 75 e9 65 d0 d2 cd 16 62 54 00 2d 00 02 01 01 00 2b 00 03 02 03 04`)
	copy(clientHello[162:194], clientPublicKey)
	c.Conn.Write(clientHello)

	// ⇇ 读 ServerHello
	serverHello := c.readRecord()
	serverPublicKey := serverHello[89 : 89+32]
	fmt.Printf("serverPublicKey: %x\n", serverPublicKey)

	// ÷ 使用 clientPrivateKey 和 serverPublicKey 计算密钥
	sharedSecret, _ := curve25519.X25519(clientPrivateKey, serverPublicKey)
	fmt.Printf("sharedSecret: %x\n", sharedSecret)

	// ÷ HKDF(sharedSecret)
	helloHash := SHA256(clientHello[5:], serverHello[5:])
	fmt.Printf("helloHash: %x\n", helloHash)
	earlySecret := HkdfExtract(hex2byte("00"), hex2byte("0000000000000000000000000000000000000000000000000000000000000000"))
	emptyHash := SHA256([]byte{})
	derivedSecret := HkdfExpandLabel(earlySecret, "derived", emptyHash, 32)
	handshakeSecret := HkdfExtract(derivedSecret, sharedSecret)
	clientHandshakeTrafficSecret := HkdfExpandLabel(handshakeSecret, "c hs traffic", helloHash, 32)
	serverHandshakeTrafficSecret := HkdfExpandLabel(handshakeSecret, "s hs traffic", helloHash, 32)
	clientHandshakeKey := HkdfExpandLabel(clientHandshakeTrafficSecret, "key", []byte{}, 16)
	serverHandshakeKey := HkdfExpandLabel(serverHandshakeTrafficSecret, "key", []byte{}, 16)
	clientHandshakeIV := HkdfExpandLabel(clientHandshakeTrafficSecret, "iv", []byte{}, 12)
	serverHandshakeIV := HkdfExpandLabel(serverHandshakeTrafficSecret, "iv", []byte{}, 12)
	fmt.Printf("clientHandshakeKey %x\n", clientHandshakeKey)
	fmt.Printf("serverHandshakeKey %x\n", serverHandshakeKey)
	fmt.Printf("clientHandshakeIV %x\n", clientHandshakeIV)
	fmt.Printf("serverHandshakeIV %x\n", serverHandshakeIV)

	// ⇇ Server Change Cipher Spec
	c.readRecord()

	// --- 明文↑ 密文↓ 分割线 ---

	// ⇇ {Encrypted Extensions, Certificate, Certificate Verify, Handshake Finished}
	wrapper := c.readRecord()
	decrypted := aes128gcmDecrypt(serverHandshakeIV, 0, serverHandshakeKey, wrapper[:5], wrapper[5:])
	fmt.Printf("decrypted %x\n", decrypted)

	// TODO: 验证对端是否确实是发送过来的证书的所有者
	// 验证方法：
	//   Certificate Verify 里包含了服务端使用证书的私钥对握手包的签名
	//   使用证书的公钥解密并和实际的握手包签名对比即可验证对端确实是发送过来的证书的所有者

	// TODO: 验证证书是否可信
	encryptedExtensionsLength := bytes2length(decrypted[1:4])
	fmt.Println(encryptedExtensionsLength)
	certficateLength := bytes2length(decrypted[4+encryptedExtensionsLength+8 : 4+encryptedExtensionsLength+8+3])
	certificate := decrypted[4+encryptedExtensionsLength+8+3 : 4+encryptedExtensionsLength+8+3+certficateLength]
	cert, _ := x509.ParseCertificate(certificate)
	fmt.Printf("Subject: %s\nIssuer: %s\nDNS Names: %v\n", cert.Subject, cert.Issuer, cert.DNSNames)

	// TODO: 验证 ServerHandshakeFinished 中的 finishedHash

	// ÷ 计算 Application Keys
	handshakeHash := SHA256(clientHello[5:], serverHello[5:], decrypted[:len(decrypted)-1])
	derivedSecret = HkdfExpandLabel(handshakeSecret, "derived", emptyHash, 32)
	fmt.Printf("%x\n", derivedSecret)
	masterSecret := HkdfExtract(derivedSecret, hex2byte(`0000000000000000000000000000000000000000000000000000000000000000`))
	fmt.Printf("%x\n", handshakeSecret)
	clientApplicationTrafficSecret := HkdfExpandLabel(masterSecret, "c ap traffic", handshakeHash, 32)
	serverApplicationTrafficSecret := HkdfExpandLabel(masterSecret, "s ap traffic", handshakeHash, 32)
	c.clientApplicationKey = HkdfExpandLabel(clientApplicationTrafficSecret, "key", []byte{}, 16)
	c.serverApplicationKey = HkdfExpandLabel(serverApplicationTrafficSecret, "key", []byte{}, 16)
	c.clientApplicationIV = HkdfExpandLabel(clientApplicationTrafficSecret, "iv", []byte{}, 12)
	c.serverApplicationIV = HkdfExpandLabel(serverApplicationTrafficSecret, "iv", []byte{}, 12)
	fmt.Printf("clientApplicationKey %x\n", c.clientApplicationKey)
	fmt.Printf("serverApplicationKey %x\n", c.serverApplicationKey)
	fmt.Printf("clientApplicationIV %x\n", c.clientApplicationIV)
	fmt.Printf("serverApplicationIV %x\n", c.serverApplicationIV)

	// ⇉ Client Change Cipher Spec
	c.Conn.Write(hex2byte(`14 03 03 00 01 01`))

	// ⇉ Client Handshake Finished
	finishedHash := handshakeHash
	finishedKey := HkdfExpandLabel(clientHandshakeTrafficSecret, "finished", []byte{}, 32)
	verifyData := HmacSHA256(finishedKey, finishedHash)
	fmt.Printf("finishedKey %x\nverifyData %x\n", finishedKey, verifyData)
	var b bytes.Buffer
	b.Write(hex2byte(`14 00 00 20`))
	b.Write(verifyData)
	b.Write([]byte{0x16})
	clientHandshakeFinished := b.Bytes()
	fmt.Printf("%x\n", clientHandshakeFinished)
	recordHeader := hex2byte(`17 03 03 00 35`)
	ciphertext := aes128gcmEncrypt(clientHandshakeIV, 0, clientHandshakeKey, recordHeader, clientHandshakeFinished)
	c.Conn.Write(recordHeader)
	c.Conn.Write(ciphertext)
}

func (c *TLS13Conn) Read() []byte {
	for {
		resp := c.readRecord()
		plaintext := aes128gcmDecrypt(c.serverApplicationIV, c.serverRecordNum, c.serverApplicationKey, resp[:5], resp[5:])
		c.serverRecordNum += 1
		if plaintext[len(plaintext)-1] == 0x17 {
			return plaintext[:len(plaintext)-1]
		} else {
			// 有可能是 New Session Ticket, etc.
			fmt.Printf("!application data: %x\n", plaintext)
		}
	}
}

func (c *TLS13Conn) Write(data []byte) {
	var b cryptobyte.Builder
	b.AddBytes([]byte{0x17, 0x03, 0x03})
	// Payload 长度 = 密文长度（=明文长度）+ 消息类型（1字节）+ 16个字节的Auth Tag（用以校验密文信息）
	b.AddUint16(uint16(len(data) + 1 + 16))
	recordHeader, _ := b.Bytes()
	fmt.Printf("record header: %x\n", recordHeader)

	payload := make([]byte, len(data)+1)
	copy(payload, data)
	// 最后一个字节是消息类型，0x17 -> Application Data
	payload[len(payload)-1] = 0x17
	fmt.Printf("payload: %x\n", payload)
	encrypted := aes128gcmEncrypt(c.clientApplicationIV, c.clientRecordNum, c.clientApplicationKey, recordHeader, payload)
	c.clientRecordNum += 1
	fmt.Printf("client application data: %x\n", encrypted)

	c.Conn.Write(append(recordHeader, encrypted...))
}

func (c *TLS13Conn) readRecord() []byte {
	// 读 Record Header（5字节）
	header := make([]byte, 5)
	_, err := io.ReadFull(c.Conn, header)
	if err != nil {
		panic(err.Error())
	}
	// 读 Record Payload
	payloadLength := (int(header[3]) << 8) + int(header[4])
	body := make([]byte, payloadLength)
	_, err = io.ReadFull(c.Conn, body)
	if err != nil {
		panic(err.Error())
	}
	return append(header, body...)
}

func bytes2length(bs []byte) (r int) {
	for _, b := range bs {
		r = (r << 8) + int(b)
	}
	return
}

func hex2byte(h string) []byte {
	b, _ := hex.DecodeString(strings.Replace(h, " ", "", -1))
	return b
}

func SHA256(a ...[]byte) []byte {
	h := sha256.New()
	for _, b := range a {
		h.Write(b)
	}
	return h.Sum(nil)
}

func HmacSHA256(key []byte, a ...[]byte) []byte {
	h := hmac.New(sha256.New, key)
	for _, b := range a {
		h.Write(b)
	}
	return h.Sum(nil)
}

func HkdfExpandLabel(secret []byte, label string, context []byte, length int) []byte {
	var hkdfLabel cryptobyte.Builder
	hkdfLabel.AddUint16(uint16(length))
	hkdfLabel.AddUint8LengthPrefixed(func(b *cryptobyte.Builder) {
		b.AddBytes([]byte("tls13 "))
		b.AddBytes([]byte(label))
	})
	hkdfLabel.AddUint8LengthPrefixed(func(b *cryptobyte.Builder) {
		b.AddBytes(context)
	})
	out := make([]byte, length)
	n, err := hkdf.Expand(sha256.New, secret, hkdfLabel.BytesOrPanic()).Read(out)
	if err != nil || n != length {
		panic("tls: HkdfExpandLabel invocation failed unexpectedly")
	}
	return out
}

func HkdfExtract(salt, secret []byte) []byte {
	return hkdf.Extract(sha256.New, secret, salt)
}

func buildIV(iv []byte, seq int) []byte {
	r := make([]byte, len(iv))
	copy(r, iv)
	for i := 0; i < 8; i++ {
		r[len(iv)-1-i] = byte(int(iv[len(iv)-1-i]) ^ ((seq >> (i * 8)) & 0xFF))
	}
	return r
}

// 返回密文+16字节的Auth Tag
func aes128gcmEncrypt(iv []byte, seq int, key, aad, plaintext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := buildIV(iv, seq)
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	return aesgcm.Seal(nil, nonce, plaintext, aad)
}

func aes128gcmDecrypt(iv []byte, seq int, key, aad, ciphertext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := buildIV(iv, seq)
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
