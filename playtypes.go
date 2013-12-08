package main

// types used by packets

type XYZ8 struct {
	X, Y, Z int8
}

type Slot struct {
	// TODO
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

type EntityMetadata struct {
	Raw []byte
}

func (d *EntityMetadata) MarshalPacket(k *PacketEncoder) {
	p := k.Get(len(d.Raw))
	if p != nil {
		copy(p, d.Raw)
	}
	k.PutUint8(0x7f)
}
func (d *EntityMetadata) UnmarshalPacket(k *PacketDecoder) {
	b := k.Get(1)
	if b == nil || b[0] == 0x7f {
		return
	}
	d.Raw = b
	for {
		if b = k.Get(1); b == nil || b[0] == 0x7f {
			return
		}
		d.Raw = d.Raw[:len(d.Raw)+1]
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
