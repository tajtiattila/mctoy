package nbt

import (
	"reflect"
	"sort"
)

func parseStructTag(s reflect.StructTag) (nam, typ string, idx int) {
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
		typ = tag[c1+1 : c2]
		if c2 < len(tag) {
			for _, ch := range tag[c2+1:] {
				if '0' <= ch && ch <= '9' {
					idx = idx*10 + int(ch-'0')
				} else {
					break
				}
			}
		}
	}
	return
}

type compoundStruct struct {
	decodeInfo map[string]int // NBT name to Go field index
	encodeInfo []fieldInfo    // Go field index to NBT info
}

type fieldInfo struct {
	kind TagKind
	name string
}

func prepareStruct(rt reflect.Type) *compoundStruct {
	cs := &compoundStruct{make(map[string]int), nil}
	m := make(map[int][]fieldInfo)
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		fn, ft, fi := parseStructTag(sf.Tag)
		if fn == "" {
			fn = sf.Name
		}
		cs.decodeInfo[fn] = i
		m[fi] = append(m[fi], fieldInfo{deduceType(ft, sf.Type), fn})
	}
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		for _, fi := range m[k] {
			cs.encodeInfo = append(cs.encodeInfo, fi)
		}
	}
	return cs
}

var knownStructs map[reflect.Type]*compoundStruct
