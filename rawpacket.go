package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrBufferShort = errors.New("Buffer insufficient for packet")
)

type PacketScanner struct {
	rd      io.Reader
	buf     []byte
	r, w    int
	lasterr error
}

func NewPacketScanner(i io.Reader) *PacketScanner {
	return &PacketScanner{rd: i, buf: make([]byte, 4)}
}

func (r *PacketScanner) SetInput(i io.Reader) {
	r.rd = i
}

// Scan() returns complete Minecraft packets
func (p *PacketScanner) Scan() (packet []byte, err error) {
	if p.w != 0 {
		copy(p.buf, p.buf[p.w:p.r])
		p.w, p.r = 0, p.r-p.w
	}
	for {
		var npkt, nread int
		if p.r != 0 {
			npkt = packetLen(p.buf[:p.r])
			if npkt != 0 && npkt < p.r {
				p.w = npkt
				return p.ret()
			}
			if p.lasterr != nil {
				// we don't have a full packet yet, so return the error, if any
				err, p.lasterr = p.lasterr, nil
				return
			}
		}
		if p.r == len(p.buf) {
			nlen := len(p.buf)*3/2 + 1
			if npkt > nlen {
				nlen = npkt*3/2 + 1
			}
			nbuf := make([]byte, nlen)
			copy(nbuf, p.buf)
			p.buf = nbuf
		}
		nread, p.lasterr = p.rd.Read(p.buf[p.r:])
		p.r += nread
	}
}

func (p *PacketScanner) ret() (packet []byte, err error) {
	packet, err, p.lasterr = p.buf[:p.w], p.lasterr, nil
	return
}

func packetLen(packet []byte) int {
	lenpkt, lenvarint := binary.Uvarint(packet)
	if lenvarint == 0 {
		return 0
	}
	return lenvarint + int(lenpkt)
}

type PacketReader struct {
	r       *bufio.Reader
	scratch []byte
	spos    int
}

func NewPacketReader(r io.Reader) *PacketReader {
	return &PacketReader{bufio.NewReader(r), make([]byte, binary.MaxVarintLen64), 0}
}

func (pr *PacketReader) Read(packet []byte) (int, error) {
	if len(packet) < binary.MaxVarintLen64 {
		return 0, ErrBufferShort
	}
	var lenp, nskip int
	for {
		var lenp0 uint64
		lenp0, nskip = binary.Uvarint(pr.scratch[:pr.spos])
		if nskip != 0 {
			lenp = int(lenp0)
			break
		}
		n, err := pr.r.Read(pr.scratch[pr.spos:binary.MaxVarintLen64])
		if err != nil {
			return 0, err
		}
		pr.spos += n
	}
	if len(packet) < lenp {
		return 0, ErrBufferShort
	}
	copy(packet, pr.scratch[nskip:pr.spos])
	pos := pr.spos - nskip
	_, err := io.ReadFull(pr.r, packet[pos:lenp])
	pr.spos = 0
	return lenp, err
}

type PacketWriter struct {
	w       *bufio.Writer
	scratch []byte
}

func NewPacketWriter(w io.Writer) *PacketWriter {
	return &PacketWriter{bufio.NewWriter(w), make([]byte, binary.MaxVarintLen64)}
}

func (pw *PacketWriter) Write(packet []byte) (int, error) {
	n := binary.PutUvarint(pw.scratch, uint64(len(packet)))
	_, err := pw.w.Write(pw.scratch[:n])
	if err != nil {
		return 0, err
	}
	nw, err := pw.w.Write(packet)
	if err != nil {
		return nw, err
	}
	return nw, pw.w.Flush()
}
