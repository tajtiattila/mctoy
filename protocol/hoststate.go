package protocol

import (
	"errors"
	"fmt"
	"reflect"
)

type HostType int

const (
	Server HostType = iota
	Client
)

func HostTypeString(h HostType) string {
	switch h {
	case Server:
		return "server"
	case Client:
		return "client"
	}
	return fmt.Sprint("invalid#", int(h))
}

// GetHostState returns the HostState for the HostType
// and CxnState specified
func GetHostState(ht HostType, c CxnState) *HostState {
	hi, ci := int(ht), int(c)
	if 0 <= ci && ci <= 3 && (hi == 0 || hi == 1) {
		return PacketData[ci][hi]
	}
	return nil
}

// HostState implements packet encoding and decoding
type HostState struct {
	ht   HostType
	xs   CxnState
	Send map[reflect.Type]PacketInfo
	Recv []*PacketInfo
}

// Encode encodes the packet into buf and returns the number of bytes used.
func (hs *HostState) Encode(buf []byte, packet interface{}) (n int, err error) {
	rv := reflect.ValueOf(packet)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if pi, ok := hs.Send[rv.Type()]; !ok {
		return 0, &ErrInvalidPacketId{
			hs.ht,
			hs.xs,
			fmt.Sprint("type invalid for state: ", rv.Type().String()),
		}
	} else {
		defer func() {
			if r := recover(); r != nil {
				n, err = 0, packetCoderError("en", rv.Type(), r)
			}
		}()
		c := MakeCoder(buf)
		c.PutVarint(pi.Id)
		pi.Write(&c, rv)
		n = c.Pos()
	}
	return
}

// Decode returns the packet encoded in buf.
func (hs *HostState) Decode(buf []byte) (p interface{}, err error) {
	c := MakeCoder(buf)
	id := c.Varint()
	pi := hs.Recv[id]
	if pi == nil {
		return nil, &ErrInvalidPacketId{
			hs.ht,
			hs.xs,
			fmt.Sprint("packet id invalid for state: ", id),
		}
	}
	pv := reflect.New(pi.Rt)
	defer func() {
		if r := recover(); r != nil {
			p, err = nil, packetCoderError("de", pi.Rt, r)
		}
	}()
	pi.Read(pv.Elem(), &c)
	p, err = pv.Interface(), nil
	return
}

var (
	PacketData [4][2]*HostState
)

type ErrInvalidPacketId struct {
	ht   HostType
	xs   CxnState
	what string
}

func (e *ErrInvalidPacketId) Error() string {
	return fmt.Sprint("mctoy-protocol: ", e.what,
		" (hosttype=", HostTypeString(e.ht),
		", state=", CxnStateString(e.xs), ")")
}

func packetCoderError(w string, rt reflect.Type, i interface{}) error {
	return errors.New(fmt.Sprint(
		"mctoy-protocol: packet ", w, "coding error: ",
		rt.Name(), ", ", i))
}
