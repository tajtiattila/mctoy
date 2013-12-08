package net

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode"
)

var (
	ErrBufferOverflow  = errors.New("Buffer overflow")
	ErrBufferUnderflow = errors.New("Buffer underflo")

	endian = binary.BigEndian
)

////////////////////////////////////////////////////////////////////////////////

type PacketMarshaler interface {
	MarshalPacket(k *PacketEncoder)
}

type PacketUnmarshaler interface {
	UnmarshalPacket(k *PacketDecoder)
}

////////////////////////////////////////////////////////////////////////////////

func tagElems(t reflect.StructTag, f func(string)) {
	ts := string(t)
	for len(ts) > 0 {
		var tag string
		if ps := strings.IndexFunc(ts, unicode.IsSpace); ps >= 0 {
			tag, ts = ts[:ps], ts[ps+1:]
		} else {
			tag, ts = ts, ""
		}
		if strings.HasPrefix(tag, "mc:") {
			tx := tag[3:]
			if len(tx) > 2 && tx[0] == '"' && tx[len(tx)-1] == '"' {
				tx = tx[1 : len(tx)-1]
				for len(tx) > 0 {
					var frag string
					if px := strings.IndexRune(tx, ','); px >= 0 {
						frag, tx = tx[:px], tx[px+1:]
					} else {
						frag, tx = tx, ""
					}
					if frag != "" {
						f(frag)
					}
				}
			}
			return
		}
	}
}

func keyVal(s string) (k, v string) {
	if p := strings.IndexRune(s, '='); p != 0 {
		return s[:p], s[p+1:]
	}
	return s, ""
}

func lenTag(t reflect.StructTag) (lt string) {
	tagElems(t, func(frag string) {
		if k, v := keyVal(frag); k == "len" {
			lt = v
		}
	})
	return
}

////////////////////////////////////////////////////////////////////////////////

type PacketDecoder struct {
	data []byte
	pos  int
	err  error
}

func MakePacketDecoder(packet []byte) PacketDecoder {
	return PacketDecoder{packet, 0, nil}
}
func (d *PacketDecoder) Error() error {
	return d.err
}
func (d *PacketDecoder) Get(size int) []byte {
	if d.pos+size <= len(d.data) {
		p := d.pos
		d.pos += size
		return d.data[p : p+size]
	}
	d.pos = len(d.data)
	d.err = ErrBufferOverflow
	return nil
}
func (d *PacketDecoder) Len() int {
	return len(d.data) - d.pos
}
func (d *PacketDecoder) Int64() int64 {
	if p := d.Get(8); p != nil {
		return int64(endian.Uint64(p))
	}
	return 0
}
func (d *PacketDecoder) Int32() int32 {
	if p := d.Get(4); p != nil {
		return int32(endian.Uint32(p))
	}
	return 0
}
func (d *PacketDecoder) Int16() int16 {
	if p := d.Get(2); p != nil {
		return int16(endian.Uint16(p))
	}
	return 0
}
func (d *PacketDecoder) Int8() int8 {
	if p := d.Get(1); p != nil {
		return int8(p[0])
	}
	return 0
}
func (d *PacketDecoder) Uint64() uint64 {
	if p := d.Get(8); p != nil {
		return endian.Uint64(p)
	}
	return 0
}
func (d *PacketDecoder) Uint32() uint32 {
	if p := d.Get(4); p != nil {
		return endian.Uint32(p)
	}
	return 0
}
func (d *PacketDecoder) Uint16() uint16 {
	if p := d.Get(2); p != nil {
		return endian.Uint16(p)
	}
	return 0
}
func (d *PacketDecoder) Uint8() uint8 {
	if p := d.Get(1); p != nil {
		return p[0]
	}
	return 0
}
func (d *PacketDecoder) Bool() bool {
	return d.Uint8() != 0
}
func (d *PacketDecoder) Varint() int {
	res, l := binary.Varint(d.data[d.pos:])
	if l == 0 {
		d.err = ErrBufferOverflow
		return 0
	}
	d.pos += l
	return int(res)
}
func (d *PacketDecoder) Uvarint() uint {
	res, l := binary.Uvarint(d.data[d.pos:])
	if l == 0 {
		d.err = ErrBufferOverflow
		return 0
	}
	d.pos += l
	return uint(res)
}
func (d *PacketDecoder) String() string {
	b := d.Get(int(d.Uvarint()))
	return string(b)
}
func (d *PacketDecoder) Float64() float64 {
	if p := d.Get(8); p != nil {
		return math.Float64frombits(endian.Uint64(p))
	}
	return 0
}
func (d *PacketDecoder) Float32() float32 {
	if p := d.Get(4); p != nil {
		return math.Float32frombits(endian.Uint32(p))
	}
	return 0
}

func (d *PacketDecoder) ArrayLength(t reflect.StructTag) int {
	switch lenTag(t) {
	case "int32div4":
		return int(d.Int32()) / 4
	case "int32":
		return int(d.Int32())
	case "int16":
		return int(d.Int16())
	case "int8":
		return int(d.Int8())
	}
	return int(d.Uvarint())
}

func (d *PacketDecoder) decodeSimple(i interface{}, tag reflect.StructTag) bool {
	switch t := i.(type) {
	case *bool:
		*t = d.Bool()
	case *int64:
		*t = d.Int64()
	case *int32:
		*t = d.Int32()
	case *int16:
		*t = d.Int16()
	case *int8:
		*t = d.Int8()
	case *uint64:
		*t = d.Uint64()
	case *uint32:
		*t = d.Uint32()
	case *uint16:
		*t = d.Uint16()
	case *uint8:
		*t = d.Uint8()
	case *int:
		*t = d.Varint()
	case *uint:
		*t = d.Uvarint()
	case *string:
		*t = d.String()
	case *float32:
		*t = d.Float32()
	case *float64:
		*t = d.Float64()
	case *[]byte:
		l := d.ArrayLength(tag)
		*t = d.Get(l)
	default:
		return false
	}
	return true
}

func (d *PacketDecoder) decodeValue(v reflect.Value, tag reflect.StructTag) {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.CanAddr() {
		i := v.Addr().Interface()
		if d.decodeSimple(i, tag) {
			return
		}
		if m, ok := i.(PacketUnmarshaler); ok {
			m.UnmarshalPacket(d)
			return
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			d.decodeValue(v.Field(i), v.Type().Field(i).Tag)
		}
	case reflect.Slice:
		l := d.ArrayLength(tag)
		et := v.Type().Elem()
		switch et.Kind() {
		case reflect.Uint8:
			v.SetBytes(d.Get(l))
		case reflect.Uint32:
			s := make([]uint32, l)
			for i := 0; i < l; i++ {
				s[i] = d.Uint32()
			}
			v.Set(reflect.ValueOf(s).Convert(v.Type()))
		default:
			av := reflect.MakeSlice(reflect.SliceOf(et), l, l)
			for i := 0; i < l; i++ {
				d.decodeValue(av.Index(i), tag)
			}
			v.Set(av)
		}
	case reflect.Array:
		l := v.Len()
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			reflect.Copy(v.Slice(0, l), reflect.ValueOf(d.Get(l)))
		default:
			for i := 0; i < l; i++ {
				d.decodeValue(v.Index(i), tag)
			}
		}
	default:
		fmt.Println(v)
		panic("can't decode")
	}
}

func (d *PacketDecoder) Decode(i interface{}) {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
	}
	d.decodeValue(v, "")
}

////////////////////////////////////////////////////////////////////////////////

type PacketEncoder struct {
	data []byte
	pos  int
	err  error
}

func MakePacketEncoder(packet []byte) PacketEncoder {
	return PacketEncoder{packet, 0, nil}
}
func (e *PacketEncoder) Error() error {
	return e.err
}
func (e *PacketEncoder) Bytes() []byte {
	return e.data[:e.pos]
}
func (e *PacketEncoder) Get(size int) []byte {
	if e.pos+size <= len(e.data) {
		p := e.pos
		e.pos += size
		return e.data[p : p+size]
	}
	e.pos = len(e.data)
	e.err = ErrBufferUnderflow
	return nil
}
func (e *PacketEncoder) PutInt64(i int64) {
	if p := e.Get(8); p != nil {
		endian.PutUint64(p, uint64(i))
	}
}
func (e *PacketEncoder) PutInt32(i int32) {
	if p := e.Get(4); p != nil {
		endian.PutUint32(p, uint32(i))
	}
}
func (e *PacketEncoder) PutInt16(i int16) {
	if p := e.Get(2); p != nil {
		endian.PutUint16(p, uint16(i))
	}
}
func (e *PacketEncoder) PutInt8(i int8) {
	if p := e.Get(1); p != nil {
		p[0] = uint8(i)
	}
}
func (e *PacketEncoder) PutUint64(i uint64) {
	if p := e.Get(8); p != nil {
		endian.PutUint64(p, i)
	}
}
func (e *PacketEncoder) PutUint32(i uint32) {
	if p := e.Get(4); p != nil {
		endian.PutUint32(p, i)
	}
}
func (e *PacketEncoder) PutUint16(i uint16) {
	if p := e.Get(2); p != nil {
		endian.PutUint16(p, i)
	}
}
func (e *PacketEncoder) PutUint8(i uint8) {
	if p := e.Get(1); p != nil {
		p[0] = i
	}
}
func (e *PacketEncoder) PutBool(b bool) {
	var i uint8
	if b {
		i = 1
	}
	e.PutUint8(i)
}
func (e *PacketEncoder) PutVarint(i int) {
	if e.pos+binary.MaxVarintLen64 <= len(e.data) {
		l := binary.PutVarint(e.data[e.pos:], int64(i))
		e.pos += l
		return
	}
	e.pos = len(e.data)
	e.err = ErrBufferUnderflow
}
func (e *PacketEncoder) PutUvarint(i uint) {
	if e.pos+binary.MaxVarintLen64 <= len(e.data) {
		l := binary.PutUvarint(e.data[e.pos:], uint64(i))
		e.pos += l
		return
	}
	e.pos = len(e.data)
	e.err = ErrBufferUnderflow
}
func (e *PacketEncoder) PutFloat64(v float64) {
	if p := e.Get(8); p != nil {
		endian.PutUint64(p, math.Float64bits(v))
	}
}
func (e *PacketEncoder) PutFloat32(v float32) {
	if p := e.Get(4); p != nil {
		endian.PutUint32(p, math.Float32bits(v))
	}
}
func (e *PacketEncoder) PutString(s string) {
	e.PutUvarint(uint(len(s)))
	if p := e.Get(len(s)); p != nil {
		copy(p, s)
	}
}

func (e *PacketEncoder) PutArrayLength(t reflect.StructTag, i int) {
	switch lenTag(t) {
	case "int32div4":
		e.PutInt32(int32(i) * 4)
	case "int32":
		e.PutInt32(int32(i))
	case "int16":
		e.PutInt16(int16(i))
	case "int8":
		e.PutInt8(int8(i))
	default:
		e.PutUvarint(uint(i))
	}
}

func (e *PacketEncoder) encodeSimple(i interface{}, tag reflect.StructTag) bool {
	switch t := i.(type) {
	case int64:
		e.PutInt64(t)
	case int32:
		e.PutInt32(t)
	case int16:
		e.PutInt16(t)
	case int8:
		e.PutInt8(t)
	case uint64:
		e.PutUint64(t)
	case uint32:
		e.PutUint32(t)
	case uint16:
		e.PutUint16(t)
	case uint8:
		e.PutUint8(t)
	case int:
		e.PutVarint(t)
	case uint:
		e.PutUvarint(t)
	case float32:
		e.PutFloat32(t)
	case float64:
		e.PutFloat64(t)
	case string:
		e.PutString(t)
	case bool:
		e.PutBool(t)
	case []byte:
		e.PutArrayLength(tag, len(t))
		if p := e.Get(len(t)); p != nil {
			copy(p, t)
		}
	default:
		return false
	}
	return true
}

func (e *PacketEncoder) encodeValue(v reflect.Value, tag reflect.StructTag) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.CanInterface() {
		if e.encodeSimple(v.Interface(), tag) {
			return
		}
	}
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(PacketMarshaler); ok {
			m.MarshalPacket(e)
			return
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			e.encodeValue(v.Field(i), v.Type().Field(i).Tag)
		}
	case reflect.Slice:
		l := v.Len()
		e.PutArrayLength(tag, l)
		et := v.Type().Elem()
		switch et.Kind() {
		case reflect.Uint8:
			if p := e.Get(l); p != nil {
				copy(p, v.Bytes())
			}
		case reflect.Uint32:
			s := v.Interface().([]uint32)
			for i := 0; i < l; i++ {
				e.PutUint32(s[i])
			}
		default:
			for i := 0; i < l; i++ {
				e.encodeValue(v.Index(i), tag)
			}
		}
	case reflect.Array:
		l := v.Len()
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			if p := e.Get(l); p != nil {
				copy(p, v.Slice(0, l).Bytes())
			}
		default:
			for i := 0; i < l; i++ {
				e.encodeValue(v.Index(i), tag)
			}
		}
	default:
		fmt.Println(v)
		panic("can't encode")
	}
}

func (e *PacketEncoder) Encode(i interface{}) {
	e.encodeValue(reflect.ValueOf(i), "")
}
