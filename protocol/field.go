package protocol

import (
	"reflect"
)

func decodeBool(v reflect.Value, c *Coder) { v.SetBool(c.Bool()) }
func encodeBool(c *Coder, v reflect.Value) { c.PutBool(v.Bool()) }

func decodeInt8(v reflect.Value, c *Coder)  { v.SetInt(int64(c.Int8())) }
func encodeInt8(c *Coder, v reflect.Value)  { c.PutInt8(int8(v.Int())) }
func decodeInt16(v reflect.Value, c *Coder) { v.SetInt(int64(c.Int16())) }
func encodeInt16(c *Coder, v reflect.Value) { c.PutInt16(int16(v.Int())) }
func decodeInt32(v reflect.Value, c *Coder) { v.SetInt(int64(c.Int32())) }
func encodeInt32(c *Coder, v reflect.Value) { c.PutInt32(int32(v.Int())) }
func decodeInt64(v reflect.Value, c *Coder) { v.SetInt(c.Int64()) }
func encodeInt64(c *Coder, v reflect.Value) { c.PutInt64(v.Int()) }

func decodeUint8(v reflect.Value, c *Coder)  { v.SetUint(uint64(c.Uint8())) }
func encodeUint8(c *Coder, v reflect.Value)  { c.PutUint8(uint8(v.Uint())) }
func decodeUint16(v reflect.Value, c *Coder) { v.SetUint(uint64(c.Uint16())) }
func encodeUint16(c *Coder, v reflect.Value) { c.PutUint16(uint16(v.Uint())) }
func decodeUint32(v reflect.Value, c *Coder) { v.SetUint(uint64(c.Uint32())) }
func encodeUint32(c *Coder, v reflect.Value) { c.PutUint32(uint32(v.Uint())) }
func decodeUint64(v reflect.Value, c *Coder) { v.SetUint(c.Uint64()) }
func encodeUint64(c *Coder, v reflect.Value) { c.PutUint64(v.Uint()) }

func decodeVarint(v reflect.Value, c *Coder)  { v.SetInt(int64(c.Varint())) }
func encodeVarint(c *Coder, v reflect.Value)  { c.PutVarint(int(v.Int())) }
func decodeVarintU(v reflect.Value, c *Coder) { v.SetUint(uint64(c.Varint())) }
func encodeVarintU(c *Coder, v reflect.Value) { c.PutVarint(int(v.Uint())) }

func decodeFloat64(v reflect.Value, c *Coder) { v.SetFloat(c.Float64()) }
func encodeFloat64(c *Coder, v reflect.Value) { c.PutFloat64(v.Float()) }
func decodeFloat32(v reflect.Value, c *Coder) { v.SetFloat(float64(c.Float32())) }
func encodeFloat32(c *Coder, v reflect.Value) { c.PutFloat32(float32(v.Float())) }

func decodeString(v reflect.Value, c *Coder) { v.SetString(c.String()) }
func encodeString(c *Coder, v reflect.Value) { c.PutString(v.String()) }
