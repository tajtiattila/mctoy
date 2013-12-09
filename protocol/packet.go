package protocol

import (
	"fmt"
)

type CxnState byte

const (
	StateHandshake CxnState = 0
	StateStatus    CxnState = 1
	StateLogin     CxnState = 2
	StatePlay      CxnState = 3
)

func CxnStateString(s CxnState) string {
	switch s {
	case StateHandshake:
		return "Handshake"
	case StateStatus:
		return "Status"
	case StateLogin:
		return "Login"
	case StatePlay:
		return "Play"
	}
	return fmt.Sprint("UnknownState#", int(s))
}

////////////////////////////////////////////////////////////////////////////////

const PktInvalid uint = ^uint(0)
