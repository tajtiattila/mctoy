package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

var (
	ErrBufferOverflow = errors.New("Buffer overflow")
	ErrUnknownArray   = errors.New("Unknown array")

	endian = binary.BigEndian
)

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
func (d *PacketDecoder) get(size int) []byte {
	if d.pos+size <= len(d.data) {
		p := d.pos
		d.pos += size
		return d.data[p : p+size]
	}
	d.pos = len(d.data)
	d.err = ErrBufferOverflow
	return nil
}
func (d *PacketDecoder) Int64() int64 {
	if p := d.get(8); p != nil {
		return int64(endian.Uint64(p))
	}
	return 0
}
func (d *PacketDecoder) Int32() int32 {
	if p := d.get(4); p != nil {
		return int32(endian.Uint32(p))
	}
	return 0
}
func (d *PacketDecoder) Int16() int16 {
	if p := d.get(2); p != nil {
		return int16(endian.Uint16(p))
	}
	return 0
}
func (d *PacketDecoder) Uint64() uint64 {
	if p := d.get(8); p != nil {
		return endian.Uint64(p)
	}
	return 0
}
func (d *PacketDecoder) Uint32() uint32 {
	if p := d.get(4); p != nil {
		return endian.Uint32(p)
	}
	return 0
}
func (d *PacketDecoder) Uint16() uint16 {
	if p := d.get(2); p != nil {
		return endian.Uint16(p)
	}
	return 0
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
	b := d.get(int(d.Uvarint()))
	println(len(b))
	return string(b)
}

func (d *PacketDecoder) ArrayLength(t reflect.StructTag) int {
	switch lenTag(t) {
	case "int32":
		return int(d.Int32())
	case "int16":
		return int(d.Int16())
	case "int":
		return int(d.Uvarint())
	}
	d.err = ErrUnknownArray
	return 0
}

func (d *PacketDecoder) decodeValue(v reflect.Value, tag reflect.StructTag) {
	switch v.Kind() {
	case reflect.Int64:
		v.SetInt(int64(d.Int64()))
	case reflect.Uint64:
		v.SetUint(uint64(d.Uint64()))
	case reflect.Int32:
		v.SetInt(int64(d.Int32()))
	case reflect.Uint32:
		v.SetUint(uint64(d.Uint32()))
	case reflect.Int16:
		v.SetInt(int64(d.Int16()))
	case reflect.Uint16:
		v.SetUint(uint64(d.Uint16()))
	case reflect.Int:
		v.SetInt(int64(d.Varint()))
	case reflect.Uint:
		v.SetUint(uint64(d.Uvarint()))
	case reflect.String:
		v.SetString(d.String())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			d.decodeValue(v.Field(i), v.Type().Field(i).Tag)
		}
	case reflect.Slice:
		l := d.ArrayLength(tag)
		et := v.Type().Elem()
		var av reflect.Value
		switch et.Kind() {
		case reflect.Uint8:
			av = reflect.ValueOf(d.get(l)).Convert(v.Type())
		case reflect.Uint32:
			s := make([]uint32, l)
			for i := 0; i < l; i++ {
				s[i] = d.Uint32()
			}
			av = reflect.ValueOf(s).Convert(v.Type())
		default:
			av = reflect.MakeSlice(et, l, l)
			for i := 0; i < l; i++ {
				d.decodeValue(av.Index(i), tag)
			}
		}
		v.Set(av)
	default:
		fmt.Println(v)
		panic("can't decode")
	}
}

func (d *PacketDecoder) Decode(i interface{}) {
	d.decodeValue(reflect.ValueOf(i).Elem(), "")
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
func (e *PacketEncoder) get(size int) []byte {
	if e.pos+size <= len(e.data) {
		p := e.pos
		e.pos += size
		return e.data[p : p+size]
	}
	e.pos = len(e.data)
	e.err = ErrBufferOverflow
	return nil
}
func (e *PacketEncoder) PutInt64(i int64) {
	if p := e.get(8); p != nil {
		endian.PutUint64(p, uint64(i))
	}
}
func (e *PacketEncoder) PutInt32(i int32) {
	if p := e.get(4); p != nil {
		endian.PutUint32(p, uint32(i))
	}
}
func (e *PacketEncoder) PutInt16(i int16) {
	if p := e.get(2); p != nil {
		endian.PutUint16(p, uint16(i))
	}
}
func (e *PacketEncoder) PutUint64(i uint64) {
	if p := e.get(8); p != nil {
		endian.PutUint64(p, i)
	}
}
func (e *PacketEncoder) PutUint32(i uint32) {
	if p := e.get(4); p != nil {
		endian.PutUint32(p, i)
	}
}
func (e *PacketEncoder) PutUint16(i uint16) {
	if p := e.get(2); p != nil {
		endian.PutUint16(p, i)
	}
}
func (e *PacketEncoder) PutVarint(i int) {
	if e.pos+binary.MaxVarintLen64 <= len(e.data) {
		l := binary.PutVarint(e.data[e.pos:], int64(i))
		e.pos += l
		return
	}
	e.pos = len(e.data)
	e.err = ErrBufferOverflow
}
func (e *PacketEncoder) PutUvarint(i uint) {
	if e.pos+binary.MaxVarintLen64 <= len(e.data) {
		l := binary.PutUvarint(e.data[e.pos:], uint64(i))
		e.pos += l
		return
	}
	e.pos = len(e.data)
	e.err = ErrBufferOverflow
}
func (e *PacketEncoder) PutString(s string) {
	e.PutUvarint(uint(len(s)))
	if p := e.get(len(s)); p != nil {
		copy(p, s)
	}
}

func (e *PacketEncoder) PutArrayLength(t reflect.StructTag, i int) {
	switch lenTag(t) {
	case "int32":
		e.PutInt32(int32(i))
	case "int16":
		e.PutInt16(int16(i))
	case "int":
		e.PutUvarint(uint(i))
	default:
		e.err = ErrUnknownArray
	}
}

func (e *PacketEncoder) encodeValue(v reflect.Value, tag reflect.StructTag) {
	switch v.Kind() {
	case reflect.Ptr:
		e.encodeValue(v.Elem(), tag)
	case reflect.Int64:
		e.PutInt64(v.Int())
	case reflect.Uint64:
		e.PutUint64(v.Uint())
	case reflect.Int32:
		e.PutInt32(int32(v.Int()))
	case reflect.Uint32:
		e.PutUint32(uint32(v.Uint()))
	case reflect.Int16:
		e.PutInt16(int16(v.Int()))
	case reflect.Uint16:
		e.PutUint16(uint16(v.Uint()))
	case reflect.Int:
		e.PutVarint(int(v.Int()))
	case reflect.Uint:
		e.PutUvarint(uint(v.Uint()))
	case reflect.String:
		e.PutString(v.String())
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
			if p := e.get(l); p != nil {
				copy(p, v.Interface().([]byte))
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
	default:
		fmt.Println(v)
		panic("can't encode")
	}
}

func (e *PacketEncoder) Encode(i interface{}) {
	e.encodeValue(reflect.ValueOf(i), "")
}
