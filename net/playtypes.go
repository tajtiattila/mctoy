package net

// types used by packets

type XYZ8 struct {
	X, Y, Z int8
}

type XYZint struct {
	X, Y, Z int
}

type Slot struct {
	ItemBlockId uint16
	ItemCount   byte
	ItemDamage  uint16
	Data        []byte // optional gzip'd NBT data
}

func (s *Slot) MarshalPacket(k *PacketEncoder) {
	k.PutUint16(s.ItemBlockId)
	if s.ItemBlockId != 0xffff {
		k.PutUint8(s.ItemCount)
		k.PutUint16(s.ItemDamage)
		if len(s.Data) != 0 {
			k.PutUint16(uint16(len(s.Data)))
			if p := k.Get(len(s.Data)); p != nil {
				copy(p, s.Data)
			}
		} else {
			k.PutUint16(0xffff)
		}
	}
}
func (s *Slot) UnmarshalPacket(k *PacketDecoder) {
	s.ItemBlockId = k.Uint16()
	if s.ItemBlockId != 0xffff {
		s.ItemCount = k.Uint8()
		s.ItemDamage = k.Uint16()
		l := int(k.Uint16())
		if l != 0 && l != 0xffff {
			s.Data = make([]byte, l)
			copy(s.Data, k.Get(l))
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

func (o *ObjectData) MarshalPacket(k *PacketEncoder) {
	k.PutUint32(o.Data)
	if o.HasSpeed {
		k.PutInt16(o.SpeedX)
		k.PutInt16(o.SpeedY)
		k.PutInt16(o.SpeedZ)
	}
}
func (o *ObjectData) UnmarshalPacket(k *PacketDecoder) {
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
	ModifierData []PropertyModifier `mc:"len=int16"` // http://www.minecraftwiki.net/wiki/Attribute#Modifiers
}

type PropertyModifier struct {
	UUID      [8]byte
	Amount    float64
	Operation byte
}

type Record struct {
	Bitmask uint32
}

type EntityData struct {
	Values []interface{}
}

func (d *EntityData) MarshalPacket(k *PacketEncoder) {
	// todo
}
func (d *EntityData) UnmarshalPacket(k *PacketDecoder) {
	d.Values = make([]interface{}, 16)
	for k.Len() > 0 {
		b := k.Uint8()
		if b == 0x7f {
			break
		}
		typ, idx := int(b&0x1f), int((b&0xe0)>>5)
		switch typ {
		case 0: // byte
			d.Values[idx] = k.Int8()
		case 1: // short
			d.Values[idx] = k.Int16()
		case 2: // int
			d.Values[idx] = k.Int32()
		case 3: // float
			d.Values[idx] = k.Float32()
		case 4: // string
			d.Values[idx] = k.String()
		case 5: // slot
			s := new(Slot)
			s.UnmarshalPacket(k)
			d.Values[idx] = s
		case 6: // x,y,z
			p := new(XYZint)
			p.X = int(k.Int32())
			p.Y = int(k.Int32())
			p.X = int(k.Int32())
			d.Values[idx] = p
		}
	}
}

type MapChunkBulkMeta struct {
	ChunkX        int32  // The X Coordinate of the chunk
	ChunkZ        int32  // The Z Coordinate of the chunk
	PrimaryBitmap uint16 // A bitmap which specifies which sections are not empty in this chunk
	AddBitmap     uint16 // A bitmap which specifies which sections need add information because of very high block ids. not yet used
}

type StatisticsEntry struct {
	Name  string // https://gist.github.com/thinkofdeath/a1842c21a0cf2e1fb5e0
	Value int    // The amount to set it to
}
