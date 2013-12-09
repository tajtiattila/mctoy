package net

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	proto "github.com/tajtiattila/mctoy/protocol"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Conn struct {
	host   string
	port   int
	c      net.Conn
	r      *bufio.Reader
	w      io.Writer
	rbuf   []byte
	wbuf   []byte
	wmtx   sync.Mutex
	state  proto.CxnState
	pkxi   [2]uint
	ht     proto.HostType // server:0 client:1
	logger *log.Logger
}

const (
	connBufLen = 65536
)

var (
	ErrPacketMismatch    = errors.New("Packet id mismatch")
	ErrServerAddrInvalid = errors.New("Server address invalid")
	ErrStateInvalid      = errors.New("State invalid")
)

func (c *Conn) Run(h PacketHandler) error {
	for {
		p, err := c.Recv()
		if err != nil {
			return err
		}

		if err = h.HandlePacket(c, p); err != nil {
			return err
		}
	}
}

const WHATPKT = true

func (c *Conn) Send(p interface{}) (err error) {
	hs := proto.GetHostState(c.ht, c.state)
	if hs == nil {
		return ErrStateInvalid
	}
	var n int
	c.wmtx.Lock()
	defer c.wmtx.Unlock()
	n, err = hs.Encode(c.wbuf, p)
	if err != nil {
		return err
	}
	if WHATPKT {
		dumpPacketId("", p, "->")
	}
	nl := binary.PutUvarint(c.wbuf[n:], uint64(n))
	_, err = c.w.Write(c.wbuf[n : n+nl])
	if err == nil {
		_, err = c.w.Write(c.wbuf[:n])
	}
	return
}

func (c *Conn) Recv() (p interface{}, err error) {
	var l uint64
	if l, err = binary.ReadUvarint(c.r); err != nil {
		return
	}
	if len(c.rbuf) < int(l) {
		c.rbuf = make([]byte, len(c.rbuf)+int(l))
	}
	b := c.rbuf[:int(l)]
	if _, err = io.ReadFull(c.r, b); err != nil {
		return
	}
	hs := proto.GetHostState(c.ht, c.state)
	if hs == nil {
		return nil, ErrStateInvalid
	}
	p, err = hs.Decode(b)
	if err != nil {
		dumpBytes(b)
	}
	if WHATPKT {
		dumpPacketId("<-", p, "")
	}
	return
}

////////////////////////////////////////////////////////////////////////////////

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
	c.rbuf = make([]byte, connBufLen)
	c.wbuf = make([]byte, connBufLen)

	c.InitIO(nil)

	c.ht = proto.Client
	c.state = proto.StateHandshake

	c.logger = log.New(os.Stdout, "cxn", log.LstdFlags)

	return nil
}

func (c *Conn) InitIO(secret []byte) {
	var r io.Reader
	r, c.w = InitPacketIO(c.c, secret)
	c.r = bufio.NewReader(r)
}

type PacketHandler interface {
	HandlePacket(c *Conn, pk interface{}) error
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

// create a PacketScanner and PacketWriter for the given io.ReadWriter,
// typically a net.Conn instance. Argument secret is used to set up
// AES/CFB8 encryption, in case it is nil, no encryption is used.
func InitPacketIO(h io.ReadWriter, secret []byte) (io.Reader, io.Writer) {
	var (
		sr io.Reader
		sw io.Writer
	)
	if secret == nil {
		sr, sw = h, h
	} else {
		aesc, err := aes.NewCipher(secret)
		if err != nil {
			panic(err)
		}
		sr = cipher.StreamReader{
			R: h,
			S: NewCFB8Decrypter(aesc, secret),
		}
		sw = cipher.StreamWriter{
			W: h,
			S: NewCFB8Encrypter(aesc, secret),
		}
	}
	if PACKETDEBUG {
		sr, sw = NewDebugReader(sr, os.Stdout), NewDebugWriter(sw, os.Stdout)
	}
	return sr, sw
}
