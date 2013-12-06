package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var (
	ErrBufferShort = errors.New("Buffer insufficient for packet")
)

func InitPacketIO(h io.ReadWriter, secret []byte) (*PacketScanner, *PacketWriter) {
	if secret == nil {
		return NewPacketScanner(h), NewPacketWriter(h)
	}
	aesc, err := aes.NewCipher(secret)
	if err != nil {
		panic(err)
	}
	var (
		sr io.Reader
		sw io.Writer
	)
	sr = cipher.StreamReader{
		R: h,
		S: NewCFB8Decrypter(aesc, secret),
	}
	sw = cipher.StreamWriter{
		W: h,
		S: NewCFB8Encrypter(aesc, secret),
	}
	if PACKETDEBUG {
		sr, sw = NewDebugReader(sr, os.Stdout), NewDebugWriter(sw, os.Stdout)
	}
	return NewPacketScanner(sr), NewPacketWriter(sw)
}

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

func (s *PacketScanner) Bytes() []byte {
	return s.buf[s.o:s.w]
}

// Scan() returns complete Minecraft packets
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
				if PACKETDEBUG {
					d := MakeDumper(os.Stdout)
					d.line("pktscan:", nil)
					d.bytes(s.Bytes())
				}
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
