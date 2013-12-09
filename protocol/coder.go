package protocol

import (
	"encoding/binary"
	"errors"
	"math"
)

var (
	ErrBufferExhausted = errors.New("Coder buffer exhausted")
	endian             = binary.BigEndian
)

// Coder is a helper class to read and write integral types for packets.
// It panics on a buffer under/overrun situation.
type Coder struct {
	data []byte
	pos  int
}

func MakeCoder(packet []byte) Coder {
	return Coder{packet, 0}
}

func (c *Coder) Pos() int {
	return c.pos
}

func (c *Coder) Len() int {
	return len(c.data) - c.pos
}

func (c *Coder) Bytes() []byte {
	return c.data[:c.pos]
}

func (c *Coder) Get(size int) []byte {
	if c.pos+size <= len(c.data) {
		p := c.pos
		c.pos += size
		return c.data[p : p+size]
	}
	panic(ErrBufferExhausted)
}
func (c *Coder) Int64() int64   { return int64(endian.Uint64(c.Get(8))) }
func (c *Coder) Int32() int32   { return int32(endian.Uint32(c.Get(4))) }
func (c *Coder) Int16() int16   { return int16(endian.Uint16(c.Get(2))) }
func (c *Coder) Int8() int8     { return int8(c.Get(1)[0]) }
func (c *Coder) Uint64() uint64 { return endian.Uint64(c.Get(8)) }
func (c *Coder) Uint32() uint32 { return endian.Uint32(c.Get(4)) }
func (c *Coder) Uint16() uint16 { return endian.Uint16(c.Get(2)) }
func (c *Coder) Uint8() uint8   { return c.Get(1)[0] }

func (c *Coder) PutInt64(i int64)   { endian.PutUint64(c.Get(8), uint64(i)) }
func (c *Coder) PutInt32(i int32)   { endian.PutUint32(c.Get(4), uint32(i)) }
func (c *Coder) PutInt16(i int16)   { endian.PutUint16(c.Get(2), uint16(i)) }
func (c *Coder) PutInt8(i int8)     { c.Get(1)[0] = uint8(i) }
func (c *Coder) PutUint64(i uint64) { endian.PutUint64(c.Get(8), i) }
func (c *Coder) PutUint32(i uint32) { endian.PutUint32(c.Get(4), i) }
func (c *Coder) PutUint16(i uint16) { endian.PutUint16(c.Get(2), i) }
func (c *Coder) PutUint8(i uint8)   { c.Get(1)[0] = i }

// Varints are unsigned in Minecraft protocol
func (c *Coder) Varint() int {
	res, l := binary.Uvarint(c.data[c.pos:])
	if l == 0 {
		panic(ErrBufferExhausted)
	}
	c.pos += l
	return int(res)
}

func (c *Coder) PutVarint(i int) {
	if c.pos+binary.MaxVarintLen64 <= len(c.data) {
		l := binary.PutUvarint(c.data[c.pos:], uint64(i))
		c.pos += l
		return
	}
	panic(ErrBufferExhausted)
}

func (c *Coder) String() string { return string(c.Get(c.Varint())) }

func (c *Coder) PutString(s string) {
	c.PutVarint(len(s))
	copy(c.Get(len(s)), s)
}

func (c *Coder) Bool() bool {
	return c.Uint8() != 0
}

func (c *Coder) PutBool(b bool) {
	var i uint8
	if b {
		i = 1
	}
	c.PutUint8(i)
}

func (c *Coder) Float64() float64 {
	return math.Float64frombits(endian.Uint64(c.Get(8)))
}
func (c *Coder) Float32() float32 {
	return math.Float32frombits(endian.Uint32(c.Get(4)))
}

func (c *Coder) PutFloat64(v float64) {
	endian.PutUint64(c.Get(8), math.Float64bits(v))
}
func (c *Coder) PutFloat32(v float32) {
	endian.PutUint32(c.Get(4), math.Float32bits(v))
}

////////////////////////////////////////////////////////////////////////////////
