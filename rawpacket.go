package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
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
	return &PacketScanner{rd: i, buf: make([]byte, 4)}
}

func NewPacketScannerSize(i io.Reader, bufsiz int) *PacketScanner {
	return &PacketScanner{rd: i, buf: make([]byte, bufsiz)}
}

func (s *PacketScanner) SetInput(i io.Reader) {
	s.rd = i
}

func (s *PacketScanner) Bytes() []byte {
	return s.buf[s.o:s.w]
}

func (s *PacketScanner) Error() error {
	return s.lasterr
}

// Scan() returns complete Minecraft packets
func (s *PacketScanner) Scan() bool {
	s.lasterr = nil
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
				return true
			}
			if s.lasterr != nil {
				// we don't have a full packet yet, but got an error
				s.o = s.w
				return false
			}
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
