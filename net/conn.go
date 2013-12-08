package net

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
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
	pkxi    [2]uint
	pkxo    PktDir // for inbound Packets, client:0 server:1
	logger  *log.Logger
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

	c.pkxo = Clientbound

	c.logger = log.New(os.Stdout, "cxn", log.LstdFlags)

	return nil
}

func (c *Conn) inboundPacketId(p Packet) uint {
	c.pkxi[0], c.pkxi[1] = p.Id()
	return c.pkxi[int(c.pkxo)]
}

func (c *Conn) outboundPacketId(p Packet) uint {
	c.pkxi[0], c.pkxi[1] = p.Id()
	return c.pkxi[1-int(c.pkxo)]
}

func (c *Conn) InitIO(secret []byte) {
	c.r, c.w = InitPacketIO(c.c, secret)
}

func (c *Conn) Send(p Packet) (err error) {
	pkc := MakePacketEncoder(c.wbuf)
	id := c.outboundPacketId(p)
	if err = CheckPacket(c.state, 1-c.pkxo, id); err != nil {
		c.logger.Print(err)
	}
	pkc.PutUvarint(id)
	pkc.Encode(p)
	if pkc.Error() != nil {
		return pkc.Error()
	}
	_, err = c.w.Write(pkc.Bytes())
	return err
}

func (c *Conn) Run(h PacketHandler) (err error) {
	var (
		p    Packet
		name string
	)
	for {
		if err = c.Scan(); err != nil {
			return
		}
		id, _ := c.PeekId()
		if p, name, err = c.Read(); err != nil {
			return
		}
		h.HandlePacket(c, id, name, p)
	}
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
	if c.inboundPacketId(p) != id {
		return ErrPacketMismatch
	}
	if err = CheckPacket(c.state, c.pkxo, id); err != nil {
		c.logger.Print(err)
	}
	pkc.Decode(p)
	if err == nil {
		if err = pkc.Error(); err != nil {
			fmt.Printf("! packet %02x: %s\n", id, err)
			dumpBytes(c.r.Bytes())
		}
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

func (c *Conn) Read() (p Packet, n string, err error) {
	id, ok := c.PeekId()
	if !ok {
		panic("missing packet id")
	}
	if p, n, err = NewPacket(c.state, c.pkxo, id); err != nil {
		return
	}
	err = c.Peek(p)
	return
}

////////////////////////////////////////////////////////////////////////////////

type PacketHandler interface {
	HandlePacket(c *Conn, pkid uint, pkname string, pk Packet) error
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
