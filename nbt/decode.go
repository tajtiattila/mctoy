package nbt

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"runtime"
)

type decoder struct {
	buf []byte
	p   int
}

var (
	ErrBufferExhausted = errors.New("Buffer exhausted")
)

func (c *decoder) Get(n int) []byte {
	if c.p+n < len(c.buf) {
		s := c.p
		c.p += n
		return c.buf[s:c.p]
	}
	panic(ErrBufferExhausted)
}

func (c *decoder) readNamed(v reflect.Value) {
}

func (c *decoder) Kind() TagKind {
	k := TagKind(c.Byte())
	if TagIntArray < k {
		panic(ErrKindUnknown{k})
	}
	return k
}

func (c *decoder) Uint(k TagKind) (v uint64) {
	if k < TagByte || TagLong < k {
		panic(errKindMismatch(k, "while reading an integer value"))
	}
	nbytes := 1 << (uint(k) - 1)
	if p := c.Get(nbytes); p != nil {
		for _, b := range p {
			v = (v << 8) | uint64(b)
		}
	}
	return
}

func (c *decoder) PutUint(k TagKind, v uint64) {
	if k < TagByte || TagLong < k {
		panic(errKindMismatch(k, "while writing an integer value"))
	}
	nbytes := 1 << (uint(k) - 1)
	if p := c.Get(nbytes); p != nil {
		shift := uint((nbytes - 1) * 8)
		for i := range p {
			p[i] = byte(v >> shift)
			shift -= 8
		}
	}
}

func (c *decoder) Float(k TagKind) (v float64) {
	switch k {
	case TagFloat:
		v = float64(math.Float32frombits(uint32(c.Uint(TagInt))))
	case TagDouble:
		v = math.Float64frombits(uint64(c.Uint(TagLong)))
	default:
		panic(errKindMismatch(k, "while reading a floating point value"))
	}
	return
}

func (c *decoder) PutFloat(k TagKind, v float64) {
	switch k {
	case TagFloat:
		c.PutUint(TagInt, uint64(math.Float32bits(float32(v))))
	case TagDouble:
		c.PutUint(TagLong, uint64(math.Float64bits(v)))
	default:
		panic(errKindMismatch(k, "while writing a floating point value"))
	}
	return
}

func (c *decoder) Byte() byte {
	if c.p < len(c.buf) {
		i := c.p
		c.p++
		return c.buf[i]
	}
	panic(ErrBufferExhausted)
}

func (c *decoder) PutByte(v byte) {
	if c.p < len(c.buf) {
		c.buf[c.p] = v
		c.p++
	}
	panic(ErrBufferExhausted)
}

func (c *decoder) String() string {
	nbytes := int(c.Uint(TagShort))
	return string(c.Get(nbytes))
}

func (c *decoder) PutString(s string) {
	nbytes := len(s)
	c.PutUint(TagShort, uint64(nbytes))
	copy(c.Get(nbytes), s)
}

func setint(rv reflect.Value, v uint64) bool {
	switch rv.Kind() {
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		rv.SetUint(v)
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		rv.SetInt(int64(v))
	default:
		return false
	}
	return true
}

func getint(rv reflect.Value) (uint64, bool) {
	switch rv.Kind() {
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		return rv.Uint(), true
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		return uint64(rv.Int()), true
	}
	return 0, false
}

func (c *decoder) skip(k TagKind) {
	switch k {
	case TagByte, TagShort, TagInt, TagLong:
		c.Uint(k)
	case TagFloat, TagDouble:
		c.Float(k)
	case TagByteArray:
		l := int(c.Uint(TagInt))
		c.Get(l)
	case TagIntArray:
		l := int(c.Uint(TagInt))
		c.Get(4 * l)
	case TagList:
		ek := c.Kind()
		l := int(c.Uint(TagInt))
		for i := 0; i < l; i++ {
			c.skip(ek)
		}
	case TagCompound:
		for {
			ek := c.Kind()
			c.String()
			c.skip(ek)
		}
	}
}

func isint(rk reflect.Kind) bool {
	switch rk {
	case reflect.Int64,
		reflect.Int32,
		reflect.Int16,
		reflect.Int8,
		reflect.Uint64,
		reflect.Uint32,
		reflect.Uint16,
		reflect.Uint8,
		reflect.Int,
		reflect.Uint:
		return true
	}
	return false
}

func (c *decoder) decode(k TagKind, rv reflect.Value) {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Interface {
		v := reflect.Zero(mapTagKinfToGo[k])
		rv.Set(v)
		rv = v
	}
	switch k {
	case TagByte, TagShort, TagInt, TagLong:
		setint(rv, c.Uint(k))
	case TagFloat, TagDouble:
		rv.SetFloat(c.Float(k))
	case TagByteArray:
		l := int(c.Uint(TagInt))
		rv.SetBytes(c.Get(l))
	case TagIntArray:
		if rv.Kind() != reflect.Slice && !isint(rv.Kind()) {
			spanic("NBT: can't decode intarray to ", rv.Type())
		}
		l := int(c.Uint(TagInt))
		if rv.Cap() < l {
			rv.Set(reflect.MakeSlice(reflect.SliceOf(rv.Type().Elem()), l, l))
		} else {
			rv.SetLen(l)
		}
		for i := 0; i < l; i++ {
			setint(rv.Index(i), c.Uint(TagInt))
		}
	case TagList:
		ek := c.Kind()
		l := int(c.Uint(TagInt))
		if rv.Cap() < l {
			reflect.MakeSlice(reflect.SliceOf(rv.Type().Elem()), l, l)
		} else {
			rv.SetLen(l)
		}
		for i := 0; i < l; i++ {
			c.decode(ek, rv.Index(i))
		}
	case TagCompound:
		c.decodeCompound(rv)
	}
}

func (c *decoder) decodeCompound(rv reflect.Value) {
	switch rv.Kind() {
	case reflect.Struct:
		c.decodeStruct(rv)
	case reflect.Map:
		c.decodeMap(rv)
	default:
		spanic("NBT: can't decode compound into ", rv.Type())
	}
}

func (c *decoder) decodeStruct(rv reflect.Value) {
	for {
		ek := c.Kind()
		if ek == TagEnd {
			break
		}
		en := c.String()
		cs := knownStructs[rv.Type()]
		if cs == nil {
			cs = prepareStruct(rv.Type())
			knownStructs[rv.Type()] = cs
		}
		if fi, ok := cs.decodeInfo[en]; ok {
			c.decode(ek, rv.Field(fi))
		} else {
			c.skip(ek)
		}
	}
}

var mapTagKinfToGo = map[TagKind]reflect.Type{
	TagByte:  reflect.TypeOf(int8(0)),
	TagShort: reflect.TypeOf(int16(0)),
	TagInt:   reflect.TypeOf(int32(0)),
	TagLong:  reflect.TypeOf(int64(0)),

	TagFloat:  reflect.TypeOf(float32(0)),
	TagDouble: reflect.TypeOf(float64(0)),

	TagString: reflect.TypeOf(""),

	TagByteArray: reflect.TypeOf(make([]byte, 1)),
	TagIntArray:  reflect.TypeOf(make([]int, 1)),

	TagList:     reflect.TypeOf(make([]interface{}, 1)),
	TagCompound: reflect.TypeOf(make(map[string]interface{})),
}

func (c *decoder) decodeMap(rv reflect.Value) {
	et := rv.Type().Elem()
	for {
		ek := c.Kind()
		if ek == TagEnd {
			break
		}
		rk := reflect.ValueOf(c.String())
		v := reflect.Zero(et)
		c.decode(ek, v)
		rv.SetMapIndex(rk, v)
	}
}

func DecodeNbt(i interface{}, b []byte) (name string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return "", InvalidUnmarshalError
	}
	c := decoder{buf: b}

	k := TagKind(c.Byte())
	if k == TagEnd {
		return "", io.EOF
	}

	name = c.String()
	c.decode(k, v)
	return name, err
}

type ErrKindUnknown struct {
	k TagKind
}

func (e *ErrKindUnknown) Error() string {
	return "ErrKindUnknown: " + e.k.String()
}

type ErrKindMismatch struct {
	k TagKind
	m string
}

func (e *ErrKindMismatch) Error() string {
	return "ErrKindMismatch: " + e.k.String() + " " + e.m
}
func errKindMismatch(k TagKind, m ...interface{}) error {
	return &ErrKindMismatch{k, fmt.Sprint(m...)}
}

type ErrTypeInvalid struct {
	t reflect.Type
	m string
}

func (e *ErrTypeInvalid) Error() string {
	return "ErrTypeInvalid: " + e.t.String() + " " + e.m
}
func errTypeInvalid(v reflect.Value, m ...interface{}) error {
	return &ErrTypeInvalid{v.Type(), fmt.Sprint(m...)}
}

var (
	ErrNotIntegral        = errors.New("Not an integral type")
	InvalidUnmarshalError = errors.New("Invalid unmarshal: type must be non-nil ptr")
)

type fieldName struct {
	Name string
}
