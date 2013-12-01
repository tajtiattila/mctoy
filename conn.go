package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type Conn struct {
	addr    string
	port    int
	c       net.Conn
	r       *PacketReader
	w       *PacketWriter
	rbuf    []byte
	wbuf    []byte
	rbufsiz int
}

const (
	bufLen = 65536
)

var (
	ErrPacketMismatch = errors.New("Packet id mismatch")
)

func Connect(addr string, port int) (*Conn, error) {
	nc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}
	c := &Conn{
		addr,
		port,
		nc,
		NewPacketReader(NewDebugReader(nc, os.Stdout)),
		NewPacketWriter(NewDebugWriter(nc, os.Stdout)),
		make([]byte, bufLen),
		make([]byte, bufLen),
		0,
	}
	return c, nil
}

type ServerStatus struct {
	Description string `json:"description"`
	Players     struct {
		Online int `json:"online"`
		Max    int `json:"max"`
	} `json:"players"`
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Ping time.Duration
}

func (s *ServerStatus) String() string {
	return fmt.Sprintf("%s %d/%d %s %s", s.Version.Name, s.Players.Online, s.Players.Max, s.Ping, s.Description)
}

func NewServerStatus(addr string, port int) (*ServerStatus, error) {
	c, err := Connect(addr, port)
	if err != nil {
		return nil, err
	}
	return c.ServerStatus()
}

func (c *Conn) ServerStatus() (*ServerStatus, error) {
	/*
		C->S : Handshake State=1
		C->S : Request
		S->C : Response
		C->S : Ping
		S->C : Ping
	*/

	err := c.Send(Handshake{
		ProtocolVersion: 4, // 1.7.2
		ServerAddress:   c.addr,
		ServerPort:      uint16(c.port),
		NextState:       StateStatus,
	})

	err = c.Send(StatusRequest{})
	if err != nil {
		return nil, err
	}

	var sr StatusResponse
	err = c.Recv(&sr)
	if err != nil {
		return nil, err
	}

	s := new(ServerStatus)
	err = json.Unmarshal([]byte(string(sr)), s)
	if err != nil {
		return nil, err
	}

	t, err := c.Ping()
	if err != nil {
		return nil, err
	}

	s.Ping = t

	return s, nil
}

func (c *Conn) Ping() (t time.Duration, err error) {
	err = c.Send(Ping{time.Now().Unix()})
	if err != nil {
		return
	}
	var p Ping
	if err = c.Recv(&p); err != nil {
		return
	}
	t = time.Now().Sub(time.Unix(p.Time, 0))
	return
}

func (c *Conn) Login(user, passwd string) error {
	/*
		C->S : Handshake State=2
		C->S : Login Start
		S->C : Encryption Key Request
		(Client Auth)
		C->S : Encryption Key Response
		(Server Auth, Both enable encryption)
		S->C : Login Success
	*/

	var ygg YggAuth
	yr, err := ygg.Authenticate(user, passwd, "")
	if err != nil {
		return err
	}
	fmt.Println(yr.AccessToken)

	err = c.Send(Handshake{
		ProtocolVersion: 4, // 1.7.2
		ServerAddress:   c.addr,
		ServerPort:      uint16(c.port),
		NextState:       StateLogin,
	})

	err = c.Send(LoginStart{yr.SelectedProfile.Name})
	if err != nil {
		return err
	}

	var erq EncryptionRequest
	if err = c.Recv(&erq); err != nil {
		return err
	}
	dumpJson(erq)

	sharedSecret := make([]byte, 16)
	if _, err = io.ReadFull(rand.Reader, sharedSecret); err != nil {
		return err
	}

	h := sha1.New()
	io.WriteString(h, erq.ServerId)
	h.Write(sharedSecret)
	h.Write(erq.PublicKey)
	sidSum := McDigest(h.Sum(nil))

	fmt.Printf("serverId: %x\n", erq.ServerId)
	fmt.Printf("sharedSecret: %x\n", sharedSecret)
	fmt.Printf("publicKey: %x\n", erq.PublicKey)
	fmt.Printf("serverIdSum: %s\n", sidSum)

	rsacipher, err := NewRSA_PKCS1v15(erq.PublicKey)
	if err != nil {
		return err
	}

	url := "https://sessionserver.mojang.com/session/minecraft/join"
	jd, err := json.Marshal(map[string]interface{}{
		"accessToken":     yr.AccessToken,
		"selectedProfile": yr.SelectedProfile,
		"serverId":        sidSum,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(jd))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return err
	}
	fmt.Printf("sessionserver ok\n")
	/*
		rsa_cipher = PKCS1_v1_5.new(RSA.importKey(pubkey))
		self.net.push(mcpacket.Packet(
			ident = (mcdata.LOGIN_STATE, mcdata.CLIENT_TO_SERVER, 0x01),
			data = {
				'shared_secret': rsa_cipher.encrypt(self.auth.shared_secret),
				'verify_token': rsa_cipher.encrypt(packet.data['verify_token']),
			}
		))
		self.net.enable_crypto(self.auth.shared_secret)
	*/
	c.Send(EncryptionResponse{
		SharedSecret: rsacipher.Encrypt(sharedSecret),
		VerifyToken:  rsacipher.Encrypt(erq.VerifyToken),
	})

	//c.EnableCrypto(sharedSecret)

	var ls LoginSuccess
	if err = c.Recv(&ls); err != nil {
		return err
	}
	dumpJson(ls)
	fmt.Println("UUID=", ls.UUID)

	return nil
}

func (c *Conn) EnableCrypto(secret []byte) {
	aescr, err := aes.NewCipher(secret)
	if err != nil {
		panic(err)
	}
	sr := cipher.NewCFBDecrypter(aescr, secret)
	c.r = NewPacketReader(NewDebugReader(cipher.StreamReader{sr, c.c}, os.Stdout))

	aescw, err := aes.NewCipher(secret)
	if err != nil {
		panic(err)
	}
	sw := cipher.NewCFBEncrypter(aescw, secret)
	c.w = NewPacketWriter(NewDebugWriter(cipher.StreamWriter{sw, c.c, nil}, os.Stdout))
}

type RSA_PKCS1v15 struct {
	pubkey *rsa.PublicKey
	err    error
}

func NewRSA_PKCS1v15(pubkey []byte) (*RSA_PKCS1v15, error) {
	k0, err := x509.ParsePKIXPublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	return &RSA_PKCS1v15{pubkey: k0.(*rsa.PublicKey)}, nil
}
func (c *RSA_PKCS1v15) Encrypt(b []byte) []byte {
	o, err := rsa.EncryptPKCS1v15(rand.Reader, c.pubkey, b)
	if err != nil {
		c.err = err
	}
	return o
}

func (c *Conn) Send(p Packet) error {
	pkc := MakePacketEncoder(c.rbuf)
	pkc.PutUvarint(p.Id())
	pkc.Encode(p)
	if pkc.Error() != nil {
		return pkc.Error()
	}
	_, err := c.w.Write(pkc.Bytes())
	return err
}

func (c *Conn) fill() (err error) {
	if c.rbufsiz == 0 {
		c.rbufsiz, err = c.r.Read(c.rbuf)
		if err != nil {
			c.rbufsiz = 0
			return err
		}
	}
	return
}

func (c *Conn) PeekId() (uint, error) {
	err := c.fill()
	if err == nil {
		return ^uint(0), err
	}
	pkc := MakePacketDecoder(c.rbuf[:c.rbufsiz])
	id := pkc.Uvarint()
	if err == nil {
		err = pkc.Error()
	}
	return id, err
}

func (c *Conn) Peek(p Packet) (err error) {
	err = c.fill()
	pkc := MakePacketDecoder(c.rbuf[:c.rbufsiz])
	id := pkc.Uvarint()
	if p.Id() != id {
		return ErrPacketMismatch
	}
	pkc.Decode(p)
	if err == nil {
		err = pkc.Error()
	}
	return
}

func (c *Conn) Pop() {
	c.rbufsiz = 0
}

func (c *Conn) Recv(p Packet) (err error) {
	err = c.Peek(p)
	c.Pop()
	return
}
