package nbt

import (
	"errors"
	"fmt"
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

Index is relevant for writing only, and is an integer to specify
the order the struct fields are written. Lower index fields are written first.
Fields with the same index are written in the order they appear in the struct
itself.

*/

type TagKind byte

func (k TagKind) String() string {
	if k <= TagIntArray {
		return tagNames[int(k)]
	}
	return "TagInvalid"
}

const (
	TagEnd       TagKind = 0
	TagByte      TagKind = 1
	TagShort     TagKind = 2
	TagInt       TagKind = 3
	TagLong      TagKind = 4
	TagFloat     TagKind = 5
	TagDouble    TagKind = 6
	TagByteArray TagKind = 7
	TagString    TagKind = 8
	TagList      TagKind = 9
	TagCompound  TagKind = 10
	TagIntArray  TagKind = 11
	TagInvalid   TagKind = 128
)

var tagNames = []string{
	"TagEnd",
	"TagByte",
	"TagShort",
	"TagInt",
	"TagLong",
	"TagFloat",
	"TagDouble",
	"TagByteArray",
	"TagString",
	"TagList",
	"TagCompound",
	"TagIntArray",
}

var mapKind = map[string]TagKind{
	"byte":      TagByte,
	"short":     TagShort,
	"int":       TagInt,
	"long":      TagLong,
	"float":     TagFloat,
	"double":    TagDouble,
	"string":    TagString,
	"list":      TagList,
	"bytearray": TagByteArray,
	"intarray":  TagIntArray,
}

func deduceKind(t string, i interface{}) TagKind {
	nn := len(t)
	for n, r := range t {
		if r == ',' {
			nn = n
			break
		}
	}
	styp := t[:nn]
	if styp != "" {
		if k, ok := mapKind[t]; ok {
			return k
		}
	} else {
		switch i.(type) {
		case uint64, *uint64, int64, *int64:
			return TagLong
		case uint32, *uint32, int32, *int32:
			return TagInt
		case uint16, *uint16, int16, *int16:
			return TagShort
		case uint8, *uint8, int8, *int8:
			return TagByte
		case uint, *uint, int, *int:
			return TagInt
		case float64, *float64:
			return TagDouble
		case float32, *float32:
			return TagFloat
		case bool, *bool:
			return TagByte
		case string, *string:
			return TagString
		case []byte, *[]byte:
			return TagByteArray
		case []uint32, *[]uint32, []int32, *[]int32:
			return TagIntArray
		}
	}
	return TagInvalid
}

func deduceTypeFromGo(rt reflect.Type) TagKind {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	switch rt.Kind() {
	case reflect.Int8, reflect.Uint8:
		return TagByte
	case reflect.Int16, reflect.Uint16:
		return TagShort
	case reflect.Int32, reflect.Uint32, reflect.Int, reflect.Uint:
		return TagInt
	case reflect.Int64, reflect.Uint64:
		return TagLong
	case reflect.Float32:
		return TagFloat
	case reflect.Float64:
		return TagDouble
	case reflect.String:
		return TagString
	case reflect.Slice:
		switch rt.Elem().Kind() {
		case reflect.Int8, reflect.Uint8:
			return TagByteArray
		case reflect.Int32, reflect.Uint32, reflect.Int, reflect.Uint:
			return TagIntArray
		}
		return TagList
	case reflect.Struct:
		return TagCompound
	case reflect.Map:
		if rt.Key().Kind() != reflect.String {
			spanic("NBT: incompatible key type", rt)
		}
		return TagCompound
	}
	spanic("NBT: unable to deduce type for ", rt)
	return 0
}

func spanic(i ...interface{}) {
	panic(errors.New(fmt.Sprint(i)))
}

func deduceType(typ string, rt reflect.Type) TagKind {
	k := TagEnd
	if typ != "" {
		var ok bool
		if k, ok = mapKind[typ]; !ok {
			spanic("NBT: unexpected type '", typ, "' in struct tag ")
		}
	} else {
		k = deduceTypeFromGo(rt)
	}

	// verify the kind we've found is correct
	if !checkTypeCompatible(rt, k) {
		spanic("NBT: Go type ", rt, " is not incompatible with NBT type ", k)
	}
	return k
}

func checkTypeCompatible(rt reflect.Type, k TagKind) bool {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	switch rt.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Uint:
		if k < TagByte || TagLong < k {
			return false
		}
	case reflect.String:
		if k != TagString {
			return false
		}
	case reflect.Float64, reflect.Float32:
		if k != TagFloat && k != TagDouble {
			return false
		}
	case reflect.Slice:
		ok := k == TagList
		switch rt.Elem().Kind() {
		case reflect.Uint8:
			ok = ok || k == TagByteArray
		case reflect.Uint, reflect.Int, reflect.Uint32, reflect.Int32:
			ok = ok || k == TagIntArray
		}
		return ok
	case reflect.Map:
		if rt.Key().Kind() != reflect.String {
			spanic("NBT: unusable Go type ", rt)
		}
		fallthrough
	case reflect.Struct:
		if k != TagCompound {
			return false
		}
	default:
		spanic("NBT: unusable Go type ", rt)
	}
	return true
}
