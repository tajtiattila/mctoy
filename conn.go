package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Conn struct {
	cfg     Config
	host    string
	port    int
	c       net.Conn
	r       *PacketScanner
	w       *PacketWriter
	wbuf    []byte
	rbufsiz int
}

const (
	bufLen = 65536
)

var (
	ErrPacketMismatch    = errors.New("Packet id mismatch")
	ErrServerAddrInvalid = errors.New("Server address invalid")
	ErrNoResponse        = errors.New("No response from server")
	ErrLoginFailed       = errors.New("Login failed")
)

func Connect(cfg Config) (*Conn, error) {

	v := strings.SplitN(cfg.Value("server"), ":", 2)
	host := v[0]
	if len(host) == 0 {
		return nil, ErrServerAddrInvalid
	}
	port := 25565
	var err error
	if len(v) > 0 {
		port, err = strconv.Atoi(v[1])
		if err != nil {
			return nil, ErrServerAddrInvalid
		}
	}

	return dial(cfg, host, port)
}

func dial(cfg Config, host string, port int) (*Conn, error) {
	nc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	var (
		sr io.Reader
		sw io.Writer
	)
	sr, sw = nc, nc
	if PACKETDEBUG {
		sr, sw = NewDebugReader(sr, os.Stdout), NewDebugWriter(sw, os.Stdout)
	}
	c := &Conn{
		cfg,
		host,
		port,
		nc,
		NewPacketScanner(sr),
		NewPacketWriter(sw),
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

func NewServerStatus(host string, port int) (*ServerStatus, error) {
	c, err := dial(nil, host, port)
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
		ServerAddress:   c.host,
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

func (c *Conn) Login(up UserPassworder) error {
	/*
		C->S : Handshake State=2
		C->S : Login Start
		S->C : Encryption Key Request
		(Client Auth)
		C->S : Encryption Key Response
		(Server Auth, Both enable encryption)
		S->C : Login Success
	*/

	var err error

	err = c.Send(Handshake{
		ProtocolVersion: 4, // 1.7.2
		ServerAddress:   c.host,
		ServerPort:      uint16(c.port),
		NextState:       StateLogin,
	})

	ygg := NewYggAuth(NewConfigStore("auth", c.cfg))
	if err = ygg.Start(up); err != nil {
		return err
	}

	err = c.Send(LoginStart{ygg.ProfileName()})
	if err != nil {
		return err
	}

	var erq EncryptionRequest
	if err = c.Recv(&erq); err != nil {
		return err
	}

	session, err := ygg.JoinSession(erq.ServerId, erq.PublicKey)
	if err != nil {
		return err
	}

	c.Send(EncryptionResponse{
		SharedSecret: session.Cipher.Encrypt(session.SharedSecret),
		VerifyToken:  session.Cipher.Encrypt(erq.VerifyToken),
	})

	c.EnableCrypto(session.SharedSecret)

	c.Scan()
	pkid, ok := c.PeekId()
	if !ok {
		return ErrNoResponse
	}

	switch pkid {
	case 0x00:
		var d Disconnect
		if err = c.Peek(&d); err != nil {
			return err
		}
		fmt.Println("Disconnect:", string(d))
		return ErrLoginFailed
	default:
		fmt.Println("Unexpected packet received, id:", pkid)
		return ErrLoginFailed
	case 0x02:
		var ls LoginSuccess
		if err = c.Peek(&ls); err != nil {
			return err
		}
		fmt.Println("Login successful")
	}

	return nil
}

func (c *Conn) EnableCrypto(secret []byte) {
	aesc, err := aes.NewCipher(secret)
	if err != nil {
		panic(err)
	}
	var (
		sr io.Reader
		sw io.Writer
	)
	sr = cipher.StreamReader{
		R: c.c,
		S: NewCFB8Decrypter(aesc, secret),
	}
	sw = cipher.StreamWriter{
		W: c.c,
		S: NewCFB8Encrypter(aesc, secret),
	}
	if PACKETDEBUG {
		sr, sw = NewDebugReader(sr, os.Stdout), NewDebugWriter(sw, os.Stdout)
	}
	c.r, c.w = NewPacketScanner(sr), NewPacketWriter(sw)
}

func (c *Conn) Send(p Packet) error {
	pkc := MakePacketEncoder(c.wbuf)
	pkc.PutUvarint(p.Id())
	pkc.Encode(p)
	if pkc.Error() != nil {
		return pkc.Error()
	}
	_, err := c.w.Write(pkc.Bytes())
	return err
}

func (c *Conn) Scan() bool {
	return c.r.Scan()
}

func (c *Conn) PeekId() (uint, bool) {
	return c.r.PacketId()
}

func (c *Conn) Peek(p Packet) (err error) {
	pkc := MakePacketDecoder(c.r.Bytes())
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

func (c *Conn) Recv(p Packet) (err error) {
	c.r.Scan()
	err = c.Peek(p)
	return
}
