package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

/*

Struct tag format is:

`nbt:"name,type,index"`

Any of which may be omitted.

Name is the structure field name by default.

Type is relevant for writing only, and the following mapping is used:

 Go Type            Default NBT type
 =======            ================
 int,uint           Int
 int64,uint64       Long
 int32,uint32       Int
 int16,uint16       Short
 int8,uint8         Byte
 string             String
 []byte             ByteArray
 []int32,[]uint32   IntArray
 []type             List
 map[string]type    Compound
 struct             Compound

Index is also relevant for writing only, and is an integer to specify
the order the struct fields are written. Fields with the same index
are written in the order they appear in the struct itself.

*/

var StructTag string

type tagCoder struct {
	buf []byte
	p   int
}

var (
	ErrBufferExhausted = errors.new("Buffer exhausted")
	ErrKindUnknown = errors.new("Kind unknown")
)

func (c *tagCoder) Get(n int) []byte {
	if c.p + n < len(c.buf) {
		s := c.p
		c.p += n
		return c.buf[s:c.p]
	}
	panic(ErrBufferExhausted)
}

func (c *tagCoder) readNamed(v reflect.Value) {
}

func (c *tagCoder) Kind() TagKind {
	k := TagKind(c.Byte())
	if TagIntArray < k {
		panic(ErrKindUnknown{k})
	}
	return k
}

func (c *tagCoder) Uint(k TagKind) (v uint64) {
	if k < TagByte || TagLong < k {
		panic(errKindMismatch(k, "while reading an integer value"))
	}
	nbytes := 1 << (uint(k)-1)
	if p := c.Get(nbytes); p != nil {
		for _, b := range p {
			v = (v<<8) | uint64(b)
		}
	}
	return
}

func (c *tagCoder) PutUint(k TagKind, v uint64) {
	if k < TagByte || TagLong < k {
		panic(errKindMismatch(k, "while writing an integer value"))
	}
	nbytes := 1 << (uint(k)-1)
	if p := c.Get(nbytes); p != nil {
		shift := uint((nbytes - 1)*8)
		for i := range p {
			p[i] = byte(v >> shift)
			shift -= 8
		}
	}
}

func (c *tagCoder) Float(k TagKind) (v float64) {
	switch k {
	case TagFloat:
		v = math.Float32frombits(uint32(c.Uint(TagInt)))
	case TagDouble:
		v = math.Float64frombits(uint64(c.Uint(TagLong)))
	default:
		panic(errKindMismatch(k, "while reading a floating point value")
	}
	return
}

func (c *tagCoder) PutFloat(k TagKind, v float64) {
	switch k {
	case TagFloat:
		c.PutUint(TagInt, uint64(math.Float32bits(v)))
	case TagDouble:
		c.PutUint(TagLong, uint64(math.Float64bits(v)))
	default:
		panic(errKindMismatch(k, "while writing a floating point value"))
	}
	return
}

func (c *tagCoder) Byte() byte {
	if c.p < len(c.buf) {
		i := c.p
		c.p++
		return c.buf[i]
	}
	panic(ErrBufferExhausted)
}

func (c *tagCoder) PutByte(v byte) {
	if c.p < len(c.buf) {
		c.buf[c.p] = v
		c.p++
	}
	panic(ErrBufferExhausted)
}

func (c *tagCoder) String() string {
	nbytes := int(c.Uint(TagShort))
	return string(c.Get(nbytes))
}

func (c *tagCoder) PutString(s string) {
	nbytes := len(s)
	c.PutUint(TagShort, uint64(nbytes))
	copy(c.Get(nbytes), s)
}

func (c *tagCoder) decodeIntegral(k TagKind, ts string, i interface{}) bool {
	if k != deduceKind(ts, i) {
		panic(errKindMismatch{k, "while parsing " + ts})
	}
	switch t := i.(type) {
	case *int64:
		*t = int64(c.Uint(k))
	case *int32:
		*t = int32(c.Uint(k))
	case *int16:
		*t = int16(c.Uint(k))
	case *int8:
		*t = int8(c.Uint(k))
	case *uint64:
		*t = uint64(c.Uint(k))
	case *uint32:
		*t = uint32(c.Uint(k))
	case *uint16:
		*t = uint16(c.Uint(k))
	case *uint8:
		*t = uint8(c.Uint(k))
	case *float64:
		*t = float64(c.Float(k))
	case *float32:
		*t = float32(c.Float(k))
	case *bool:
		*t = 0 != c.Uint(k)
	case *string:
		*t = c.String()
	case *[]byte:
		if k != TagByteArray {
			panic(errKindMismatch(k, "while reading a byte slice"))
		}
		l := int(c.Uint(TagInt))
		*t = c.Get(l)
	case *[]int32:
		if k != TagIntArray {
			panic(errKindMismatch(k, "while reading an int32 slice"))
		}
		l := int(c.Uint(TagInt))
		*t = make([]int32, l)
		for i := range *t {
			(*t)[i] = int32(c.Uint(TagInt))
		}
	case *[]uint32:
		if k != TagIntArray {
			panic(errKindMismatch(k, "while reading an uint32 slice"))
		}
		l := int(c.Uint(TagInt))
		*t = make([]uint32, l)
		for i := range *t {
			(*t)[i] = uint32(c.Uint(TagInt))
		}
	default:
		return false
	}
	return true
}

func (c *tagCoder) encodeIntegral(t string, i interface{}) error {
	k := deduceKind(t, i)
	switch t := i.(type) {
	case int64:
		c.PutUint(k, uint64(t))
	case int32:
		c.PutUint(k, uint64(t))
	case int16:
		c.PutUint(k, uint64(t))
	case int8:
		c.PutUint(k, uint64(t))
	case uint64:
		c.PutUint(k, uint64(t))
	case uint32:
		c.PutUint(k, uint64(t))
	case uint16:
		c.PutUint(k, uint64(t))
	case uint8:
		c.PutUint(k, uint64(t))
	case float64:
		c.PutFloat(k, float64(t))
	case float32:
		c.PutFloat(k, float64(t))
	case bool:
		var i uint64
		if t {
			i = 1
		}
		c.PutUint(k, i)
	case string:
		c.PutString(t)
	case []byte:
		l := len(t)
		c.PutUint(TagInt, uint64(l))
		if p := c.Get(l); p == nil {
			copy(p, t)
		}
	case []int32:
		l := len(t)
		c.PutUint(TagInt, uint64(l))
		for _, v := range t {
			c.PutUint(TagInt, uint64(v))
		}
	case []uint32:
		l := len(t)
		c.PutUint(TagInt, uint64(l))
		for _, v := range t {
			c.PutUint(TagInt, uint64(v))
		}
	default:
		return ErrNotIntegral
	}
	return nil
}

func isKindIntegral(k TagKind) bool {
	switch k {
    case TagByte,
			TagShort,
			TagInt,
			TagLong,
			TagFloat,
			TagDouble,
			TagByteArray,
			TagString,
			TagIntArray:
		return true
	}
	return false
}

func setint(rv reflect.Value, v uint64) bool {
	switch rv.Kind() {
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		rv.SetUint(v)
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		rv.SetInt(uint64(v))
	default:
		return false
	}
	return true
}

func getint(rv reflect.Value) (v uint64, bool) {
	switch rv.Kind() {
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		return rv.GetUint(), true
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		return uint64(rv.GetInt()), true
	}
	return false
}

func (c *tagCoder) decode(k TagKind, rv reflect.Value, fn *string) {
	if fn != nil {
		*fn = c.String()
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if isKindIntegral(k) {
		err = decodeIntegral(k, "", rv.Addr().Interface())
		if err != ErrNotIntegral {
			return
		}
		err = nil
	}
	switch k {
	case TagByte, TagShort, TagInt, TagLong:
		setint(rv, c.Uint(k))
	case TagFloat, TagDouble:
		rv.SetFloat(c.Float(k))
	case TagByteArray:
		l := int(c.Uint(TagInt))
		rv.SetBytes(d.Get(l))
	case TagIntArray:
		if rv.Kind() != reflect.Slice && !isint(rv.Type().Elem) {
			panic(errTypeMismatch())
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
			decode(ek, rv.Index(i), nil)
		}
	case TagCompound:
		for {
			ek := c.Kind()
			nam := c.String()
		}
	}
}

func typesUncompatible(rt reflect.Type, k TagKind) {
	panic(errors.New("NBT: Go type " + rt.String() +
		" is not incompatible with NBT type " + k.String()))
}

func deduceType(typ string, rt reflect.Type) TagKind {
	k := TagEnd
	if typ != "" {
		var ok bool
		if k, ok = mapKind[typ]; !ok {
			panic(errors.New("NBT: unexpected type '" + typ + "' in struct tag "))
		}
	} else {
	}

	// verify the kind we've found
	switch rt.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Uint:
		if k < TagByte || TagLong < k {
			typesUncompatible(rt, k)
		}
	case reflect.String:
		if k != TagString {
			typesUncompatible(rt, k)
		}
	case reflect.Float64, reflect.Float32:
		if k != TagFloat && k != TagDouble {
			typesUncompatible(rt, k)
		}
	case reflect.Slice:
		compat := k == TagList
		switch reflect.Elem().Kind() {
			case reflect.Byte:
				compat |= k == TagByteArray
			case reflect.Uint, reflect.Int, reflect.Uint32, reflect.Int32:
				compat |= k == TagIntArray
		}
		compat :=
			(rt.Elem().Kind() == reflect.Byte && k == TagByteArray) ||
			(rt.Elem().Kind
		if !compat {
			typesUncompatible(rt, k)
		}
	case reflect.Struct, reflect.Map:
		if k != TagCompound {
			typesUncompatible(rt, k)
		}
	default:
	}
}

func parseStructTag(s reflect.StructTag) (nam typ string, idx int) {
	tag := s.Get("nbt")
	c1, c2 := len(tag), len(tag)
	for n, ch := range tag {
		if ch == ',' {
			if n < c1 {
				c1 = n
			} else {
				c2 = n
				break
			}
		}
	}
	nam = tag[:c1]
	if c1 < len(tag) {
		typ = parseType(tag[c1+1:c2])
		if c2 < len(tag) {
			for _, ch := range(tag[c2+1:]) {
				if '0' <= ch && ch <= '9' {
					idx = idx * 10 + int(ch-'0')
				} else {
					break
				}
			}
		}
	}
	return
}

func prepareStruct(rv reflect.Value) {
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		fn, ft, fi := parseStructTag(sf.Tag)
	}
}

func (c *tagCoder) decodeCompound() {
	n := 0
	for {
		ek := c.Kind()
		if er == TagEnd {
			break
		}
		en := c.String()
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
		return InvalidUnmarshalError
	}
	c := tagCoder{buf:b}

	k := TagKind(c.Byte())
	if k == TagEnd {
		return "", io.EOF
	}

	c.decode(v, &name)
	return name, c.err
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
	ErrNotIntegral = errors.New("Not an integral type")
	InvalidUnmarshalError = errors.New("Invalid unmarshal: type must be non-nil ptr")
)

type fieldName struct {
	Name string
}

type tagCoderCtx struct {
	v     reflect.Value
	named bool
	name  string
}

func (x *tagCoderCtx) initDecode(i interface{}) {
	x.v = reflect.ValueOf(i)
	x.prepareDecode()
	x.named = true
	x.name = ""
}

func (x *tagCoderCtx) prepareDecode() {
	for x.v.Kind() == reflect.Ptr {
		if x.v.IsNil() {
			x.v.Set(reflect.New(v.Type().Elem()))
		}
		x.v = x.v.Elem()
	}
}

func (x *tagCoder) decIntf() interface{} {
	if x.i != nil {
		if x.v.Kind() != reflect.Ptr {
			x.v = x.v.Addr()
		}
		x.i = x.v.Interface()
	}
}

func (x *tagCoder) decValue() {
	if !x.v.IsValid() {
		x.v = reflect.ValueOf(x.i)
	}

	if x.v.CanAddr() {
		i := v.Addr().Interface()
		if d.decodeSimple(i, tag) {
			return
		}
		if m, ok := i.(PacketUnmarshaler); ok {
			m.UnmarshalPacket(d)
			return
		}
	}
}






