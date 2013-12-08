package main

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
