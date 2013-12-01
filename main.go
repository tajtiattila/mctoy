package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

type Packet struct {
}

type Conn struct {
	r       bufio.Reader
	w       bufio.Writer
	scratch []byte
	spos    int
}

func NewConn() *Conn {
	return &Conn{scratch: make([]byte, 32)}
}

type PacketHandler interface {
	Handle([]byte) error
}

func ReadPackets(r0 io.Reader, h PacketHandler) error {
	scratch := make([]byte, 16)
	packetdata := make([]byte, 65536)
	r := bufio.NewReader(r0)
	for {
		var (
			plen   int64
			pos, l int
		)
		for l == 0 {
			n, err := r.Read(scratch[pos:binary.MaxVarintLen64])
			if err != nil {
				return err
			}
			pos += n
			plen, l = binary.Varint(scratch[:n])
		}
		lenp := int(plen)
		if len(packetdata) < lenp {
			packetdata = make([]byte, len(packetdata)*2+lenp)
		}
		copy(packetdata, scratch[l:pos])
		pos -= l
		_, err := io.ReadFull(r, packetdata[pos:lenp])
		if err != nil {
			return err
		}
		err = h.Handle(packetdata[:lenp])
		if err != nil {
			return err
		}
	}
}

/*
func (b *Bot) Handle(p []byte) {
	r := bytes.NewReader(p)
	pkid, _ := binary.ReadVarint(r)
	switch pkid {
	case 0x00:
		b.Respond(p)
	}
}
*/
