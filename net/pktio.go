package net

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
	r, w, o int
	lasterr error
}

const DefaultPacketBufferSize = 4096

func NewPacketScanner(i io.Reader) *PacketScanner {
	return NewPacketScannerSize(i, DefaultPacketBufferSize)
}

func NewPacketScannerSize(i io.Reader, bufsiz int) *PacketScanner {
	return &PacketScanner{rd: i, buf: make([]byte, bufsiz)}
}

// Bytes returns the packet found by Scan. The returned
// byte slice is only valid until the next call to Scan.
func (s *PacketScanner) Bytes() []byte {
	return s.buf[s.o:s.w]
}

// Scan reads the next complete Minecraft packet.
func (s *PacketScanner) Scan() error {
	if s.w != 0 {
		copy(s.buf, s.buf[s.w:s.r])
		s.w, s.r = 0, s.r-s.w
	}
	for {
		var epkt, nread int
		if s.r != 0 {
			npkt0, nl := binary.Uvarint(s.buf[:s.r])
			epkt = nl + int(npkt0)
			if nl != 0 && epkt <= s.r {
				s.o, s.w = nl, epkt
				return nil
			}
		}
		if s.lasterr != nil {
			// we don't have a full packet yet, but got an error
			s.o = s.w
			return s.lasterr
		}
		if s.r == len(s.buf) {
			nlen := len(s.buf)*3/2 + 1
			if epkt > nlen {
				nlen = epkt*3/2 + 1
			}
			nbuf := make([]byte, nlen)
			copy(nbuf, s.buf)
			s.buf = nbuf
		}
		nread, s.lasterr = s.rd.Read(s.buf[s.r:])
		s.r += nread
	}
}

func (s *PacketScanner) PacketId() (uint, bool) {
	i, ni := binary.Uvarint(s.Bytes()) // packet length
	if ni != 0 {
		return uint(i), true
	}
	return 0, false
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
