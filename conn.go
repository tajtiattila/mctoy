package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Conn struct {
	host    string
	port    int
	c       net.Conn
	r       *PacketScanner
	w       *PacketWriter
	wbuf    []byte
	rbufsiz int
	state   CxnState
}

const (
	connBufLen = 65536
)

var (
	ErrPacketMismatch    = errors.New("Packet id mismatch")
	ErrServerAddrInvalid = errors.New("Server address invalid")
)

func (c *Conn) dial(addr string) error {

	v := strings.SplitN(addr, ":", 2)
	c.host = v[0]
	if c.host == "" {
		return ErrServerAddrInvalid
	}
	c.port = 25565
	var err error
	if len(v) > 0 {
		c.port, err = strconv.Atoi(v[1])
		if err != nil {
			return ErrServerAddrInvalid
		}
	}

	c.c, err = net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return err
	}
	c.wbuf = make([]byte, connBufLen)

	c.InitIO(nil)

	c.state = StateHandshake

	return nil
}

func (c *Conn) InitIO(secret []byte) {
	c.r, c.w = InitPacketIO(c.c, secret)
}

func (c *Conn) Send(p Packet) error {
	pkc := MakePacketEncoder(c.wbuf)
	pkc.PutUvarint(p.Id(PktDisp{S: c.state, D: Serverbound}))
	pkc.Encode(p)
	if pkc.Error() != nil {
		return pkc.Error()
	}
	_, err := c.w.Write(pkc.Bytes())
	return err
}

func (c *Conn) Scan() error {
	return c.r.Scan()
}

func (c *Conn) PeekId() (uint, bool) {
	return c.r.PacketId()
}

func (c *Conn) Peek(p Packet) (err error) {
	pkc := MakePacketDecoder(c.r.Bytes())
	id := pkc.Uvarint()
	if p.Id(PktDisp{S: c.state, D: Clientbound}) != id {
		return ErrPacketMismatch
	}
	pkc.Decode(p)
	if err == nil {
		err = pkc.Error()
	}
	return
}

func (c *Conn) Recv(p Packet) (err error) {
	err = c.r.Scan()
	if err == nil {
		err = c.Peek(p)
	}
	return
}

////////////////////////////////////////////////////////////////////////////////

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
