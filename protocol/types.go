package protocol

import (
	"reflect"
)

// types used by packets

type XYZ8 struct {
	X, Y, Z int8
}

type XYZint struct {
	X, Y, Z int
}

type Slot struct {
	Id     uint16
	Count  byte
	Damage uint16
	Tag    []byte // optional gzip'd NBT data
}

func (s *Slot) MarshalPacket(k *Coder) {
	k.PutUint16(s.Id)
	if s.Id != 0xffff {
		k.PutUint8(s.Count)
		k.PutUint16(s.Damage)
		if len(s.Tag) != 0 {
			k.PutUint16(uint16(len(s.Tag)))
			if p := k.Get(len(s.Tag)); p != nil {
				copy(p, s.Tag)
			}
		} else {
			k.PutUint16(0xffff)
		}
	}
}
func (s *Slot) UnmarshalPacket(k *Coder) {
	s.Id = k.Uint16()
	if s.Id != 0xffff {
		s.Count = k.Uint8()
		s.Damage = k.Uint16()
		l := int(k.Uint16())
		if l != 0 && l != 0xffff {
			s.Tag = make([]byte, l)
			copy(s.Tag, k.Get(l))
		}
	}
}

type ObjectData struct {
	Data     uint32
	HasSpeed bool
	SpeedX   int16
	SpeedY   int16
	SpeedZ   int16
}

func (o *ObjectData) MarshalPacket(k *Coder) {
	k.PutUint32(o.Data)
	if o.HasSpeed {
		k.PutInt16(o.SpeedX)
		k.PutInt16(o.SpeedY)
		k.PutInt16(o.SpeedZ)
	}
}
func (o *ObjectData) UnmarshalPacket(k *Coder) {
	o.Data = k.Uint32()
	if o.HasSpeed = k.Len() > 0; o.HasSpeed {
		o.SpeedX = k.Int16()
		o.SpeedY = k.Int16()
		o.SpeedZ = k.Int16()
	}
}

type PropertyData struct {
	Key          string
	Value        float64
	ModifierData []PropertyModifier `mc:"len=short"` // http://www.minecraftwiki.net/wiki/Attribute#Modifiers
}

type PropertyModifier struct {
	UUID      [16]byte
	Amount    float64
	Operation byte
}

type Record uint32 // Bitmask

type Metadata map[int]interface{}

func (d Metadata) MarshalPacket(k *Coder) {
	for i, v := range d {
		b := k.Get(1)
		var typ byte
		switch t := v.(type) {
		case int8:
			typ = 0
			k.PutInt8(t)
		case int16:
			typ = 1
			k.PutInt16(t)
		case int32:
			typ = 2
			k.PutInt32(t)
		case float32:
			typ = 3
			k.PutFloat32(t)
		case string:
			typ = 4
			k.PutString(t)
		case *Slot:
			typ = 5
			t.MarshalPacket(k)
		case *XYZint:
			typ = 6
			k.PutInt32(int32(t.X))
			k.PutInt32(int32(t.Y))
			k.PutInt32(int32(t.Z))
		}
		b[0] = (typ << 5) | byte(i&0x1f)
	}
	k.PutUint8(0x7f)
}

func (d Metadata) UnmarshalPacket(k *Coder) {
	d = make(Metadata)
	for k.Len() > 0 {
		b := k.Uint8()
		if b == 0x7f {
			break
		}
		typ, idx := int((b&0xe0)>>5), int(b&0x1f)
		switch typ {
		case 0: // byte
			d[idx] = k.Int8()
		case 1: // short
			d[idx] = k.Int16()
		case 2: // int
			d[idx] = k.Int32()
		case 3: // float
			d[idx] = k.Float32()
		case 4: // string
			d[idx] = k.String()
		case 5: // slot
			s := new(Slot)
			s.UnmarshalPacket(k)
			d[idx] = s
		case 6: // x,y,z
			p := new(XYZint)
			p.X = int(k.Int32())
			p.Y = int(k.Int32())
			p.X = int(k.Int32())
			d[idx] = p
		}
	}
}

type MapChunkBulkMeta struct {
	ChunkX        int32  // The X Coordinate of the chunk
	ChunkZ        int32  // The Z Coordinate of the chunk
	PrimaryBitmap uint16 // A bitmap which specifies which sections are not empty in this chunk
	AddBitmap     uint16 // A bitmap which specifies which sections need add information because of very high block ids. not yet used
}

var mapChunkBulkMetaCoder = cacheType(reflect.TypeOf(MapChunkBulkMeta{}))

type StatisticsEntry struct {
	Name  string // https://gist.github.com/thinkofdeath/a1842c21a0cf2e1fb5e0
	Value int    // The amount to set it to
}
