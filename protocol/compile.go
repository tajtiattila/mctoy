package protocol

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type PacketMarshaler interface {
	MarshalPacket(k *Coder)
}

type PacketUnmarshaler interface {
	UnmarshalPacket(k *Coder)
}

type typeCoder struct {
	rf ReadFunc
	wf WriteFunc
}

func (tc *typeCoder) valid() bool {
	return tc.rf != nil && tc.wf != nil
}

var (
	structInfoMutex sync.RWMutex
	structInfoMap   = map[reflect.Type]typeCoder{}
)

func cacheType(rt reflect.Type) typeCoder {
	structInfoMutex.RLock()
	tc, ok := structInfoMap[rt]
	structInfoMutex.RUnlock()

	if ok {
		return tc
	}

	tc = compileType(tagMap{}, rt)

	structInfoMutex.Lock()
	structInfoMap[rt] = tc
	structInfoMutex.Unlock()

	return tc
}

type tagMap map[string]string

func decodeTag(st reflect.StructTag) tagMap {
	m := make(tagMap)
	for _, part := range strings.Split(st.Get("mc"), ",") {
		if part == "" {
			continue
		}
		psep := strings.IndexRune(part, '=')
		if psep == -1 {
			panic(errors.New("Invalid struct tag: " + string(st)))
		}
		n, v := part[:psep], part[psep+1:]
		m[n] = v
	}
	return m
}

func isUnsigned(rt reflect.Type) bool {
	switch rt.Kind() {
	case reflect.Int,
		reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return false
	case reflect.Uint,
		reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return true
	}
	panic(errors.New("isUnsigned called for non-integer"))
}

func makeIntCoder(tags tagMap, rt reflect.Type) (tc typeCoder) {
	t := tags["type"]
	if t == "" {
		switch rt.Kind() {
		case reflect.Int, reflect.Uint:
			t = "varint"
		case reflect.Int64, reflect.Uint64:
			t = "long"
		case reflect.Int32, reflect.Uint32:
			t = "int"
		case reflect.Int16, reflect.Uint16:
			t = "short"
		case reflect.Int8, reflect.Uint8:
			t = "byte"
		}
	}
	tc = makeIntCoderSub(t, isUnsigned(rt))
	if !tc.valid() {
		panic(errors.New("Unrecognised int type: " + rt.String() + " " + t))
	}
	return
}

func makeIntCoderSub(t string, unsigned bool) (tc typeCoder) {
	switch t {
	case "varint":
		if unsigned {
			tc = typeCoder{decodeVarintU, encodeVarintU}
		} else {
			tc = typeCoder{decodeVarint, encodeVarint}
		}
	case "long":
		if unsigned {
			tc = typeCoder{decodeUint64, encodeUint64}
		} else {
			tc = typeCoder{decodeInt64, encodeInt64}
		}
	case "int":
		if unsigned {
			tc = typeCoder{decodeUint32, encodeUint32}
		} else {
			tc = typeCoder{decodeInt32, encodeInt32}
		}
	case "short":
		if unsigned {
			tc = typeCoder{decodeUint16, encodeUint16}
		} else {
			tc = typeCoder{decodeInt16, encodeInt16}
		}
	case "byte":
		if unsigned {
			tc = typeCoder{decodeUint8, encodeUint8}
		} else {
			tc = typeCoder{decodeInt8, encodeInt8}
		}
	}
	return
}

func makeFloatCoder(tags tagMap, rt reflect.Type) (tc typeCoder) {
	t := tags["type"]
	if t == "" {
		switch rt.Kind() {
		case reflect.Float32:
			t = "float"
		case reflect.Float64:
			t = "double"
		}
	}
	switch t {
	case "float":
		tc = typeCoder{decodeFloat32, encodeFloat32}
	case "double":
		tc = typeCoder{decodeFloat64, encodeFloat64}
	default:
		panic(errors.New("Unrecognised float type: " + rt.String() + " " + t))
	}
	return
}

func makeSliceCoder(tags tagMap, rt reflect.Type) (tc typeCoder) {
	l, d := "", 1
	var ok bool
	if l, ok = tags["len"]; !ok {
		l = "varint"
	}
	if sd, ok := tags["div"]; ok {
		var err error
		if d, err = strconv.Atoi(sd); err != nil {
			panic(err)
		}
		if d == 0 {
			panic(errors.New("slice length division by zero"))
		}
	}
	lenc := makeIntCoderSub(l, false)
	if !lenc.valid() {
		panic(errors.New("Length coder is nil"))
	}
	if rt.Elem().Kind() == reflect.Uint8 {
		if d != 1 {
			panic(errors.New("byte slice length division must be 1"))
		}
		tc.rf = func(v reflect.Value, c *Coder) {
			var l int
			lenc.rf(reflect.ValueOf(&l).Elem(), c)
			v.SetBytes(c.Get(l))
		}
		tc.wf = func(c *Coder, v reflect.Value) {
			b := v.Bytes()
			l := len(b)
			lenc.wf(c, reflect.ValueOf(l))
			copy(c.Get(l), b)
		}
	} else {
		rte := rt.Elem()
		elc := compileType(tags, rte)
		if !elc.valid() {
			panic("can't en/decode slice element: " + rte.String())
		}
		tc.rf = func(v reflect.Value, c *Coder) {
			var l int
			lenc.rf(reflect.ValueOf(&l).Elem(), c)
			l /= d
			v.Set(reflect.MakeSlice(rt, l, l))
			for i := 0; i < l; l++ {
				elc.rf(v.Index(i), c)
			}
		}
		tc.wf = func(c *Coder, v reflect.Value) {
			l := v.Len() * d
			lenc.wf(c, reflect.ValueOf(l))
			for i := 0; i < l; l++ {
				elc.wf(c, v.Index(i))
			}
		}
	}
	return
}

func makeArrayCoder(tags tagMap, rt reflect.Type) (tc typeCoder) {
	if rt.Elem().Kind() != reflect.Uint8 {
		panic(errors.New("only byte arrays are supported"))
	}
	l := rt.Len()
	tc.rf = func(v reflect.Value, c *Coder) {
		b := v.Slice(0, l).Bytes()
		copy(b, c.Get(l))
	}
	tc.wf = func(c *Coder, v reflect.Value) {
		b := v.Slice(0, l).Bytes()
		copy(c.Get(l), b)
	}
	return
}

func compileType(tags tagMap, rt reflect.Type) (tc typeCoder) {
	var tci typeCoder
	defer func() {
		if r := recover(); r != nil {
			var s, sep string
			if tci.rf != nil {
				s, sep = "PacketUnmarshaler", ", "
			}
			if tci.wf != nil {
				s += sep + "PacketMarshaler"
			}
			if s != "" {
				s = " (" + s + ")"
			}
			panic(errors.New(fmt.Sprint("compileType: ", rt.String(), s)))
		}
	}()
	rv := reflect.New(rt)
	pinst, inst := rv.Interface(), rv.Elem().Interface()
	if _, ok := inst.(PacketUnmarshaler); ok {
		tci.rf = func(v reflect.Value, c *Coder) {
			x := v.Interface().(PacketUnmarshaler)
			x.UnmarshalPacket(c)
		}
	} else {
		if _, ok := pinst.(PacketUnmarshaler); ok {
			tci.rf = func(v reflect.Value, c *Coder) {
				x := v.Addr().Interface().(PacketUnmarshaler)
				x.UnmarshalPacket(c)
			}
		}
	}
	if _, ok := inst.(PacketMarshaler); ok {
		tci.wf = func(c *Coder, v reflect.Value) {
			x := v.Interface().(PacketMarshaler)
			x.MarshalPacket(c)
		}
	} else {
		if _, ok := pinst.(PacketMarshaler); ok {
			tci.wf = func(c *Coder, v reflect.Value) {
				x := v.Addr().Interface().(PacketMarshaler)
				x.MarshalPacket(c)
			}
		}
	}

	if tci.rf == nil || tci.wf == nil {
		switch rt.Kind() {
		case reflect.Bool:
			tc = typeCoder{decodeBool, encodeBool}
		case reflect.Int, reflect.Uint,
			reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			tc = makeIntCoder(tags, rt)
		case reflect.Float64, reflect.Float32:
			tc = makeFloatCoder(tags, rt)
		case reflect.String:
			tc = typeCoder{decodeString, encodeString}
		case reflect.Slice:
			tc = makeSliceCoder(tags, rt)
		case reflect.Array:
			tc = makeArrayCoder(tags, rt)
		case reflect.Struct:
			tc = compileStruct(rt)
		}
	}

	if tci.rf != nil {
		tc.rf = tci.rf
	}
	if tci.wf != nil {
		tc.wf = tci.wf
	}

	testCoder(rt, tc)
	return
}

func compileStruct(rt reflect.Type) typeCoder {
	var fields []typeCoder
	for i := 0; i < rt.NumField(); i++ {
		tc := compileField(rt, i)

		fields = append(fields, tc)
	}
	return typeCoder{decodeStruct(fields), encodeStruct(fields)}
}

func compileField(rt reflect.Type, i int) typeCoder {
	sf := rt.Field(i)
	defer func() {
		if r := recover(); r != nil {
			panic(errors.New("compileField: " +
				rt.String() + "." + sf.Name + " (" + sf.Type.String() + ")"))
		}
	}()
	return compileType(decodeTag(sf.Tag), sf.Type)
}

// testCoder tests if we can encode/decode the value
func testCoder(rt reflect.Type, tc typeCoder) {
	s := "encode"
	defer func() {
		if r := recover(); r != nil {
			panic(errors.New("typeCoder can't " + s + " type: " + rt.String()))
		}
	}()
	v := reflect.New(rt).Elem()
	cw := &Coder{data: make([]byte, 64)}
	tc.wf(cw, v)
	s = "decode"
	cr := &Coder{data: cw.Bytes()}
	tc.rf(v, cr)
}

func decodeStruct(fields []typeCoder) ReadFunc {
	return func(v reflect.Value, c *Coder) {
		for i, f := range fields {
			f.rf(v.Field(i), c)
		}
	}
}
func encodeStruct(fields []typeCoder) WriteFunc {
	return func(c *Coder, v reflect.Value) {
		for i, f := range fields {
			f.wf(c, v.Field(i))
		}
	}
}
