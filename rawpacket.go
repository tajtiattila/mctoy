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
