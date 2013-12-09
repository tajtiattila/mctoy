package protocol

// StateHandshake
////////////////////////////////////////////////////////////////////////////////

type Handshake struct {
	// 0x00 ->Server
	ProtocolVersion uint
	ServerAddress   string
	ServerPort      uint16
	NextState       uint
}

func (h *Handshake) StateUpdate() CxnState { return CxnState(h.NextState) }

// StateStatus
////////////////////////////////////////////////////////////////////////////////

// 0x00 ->Server
type StatusRequest struct{}

// 0x00 ->Client
type StatusResponse struct {
	JSON string
}

// 0x01 ->Server ->Client
type StatusPing struct {
	Time int64
}

// StateLogin
////////////////////////////////////////////////////////////////////////////////

// 0x00 ->Server
type LoginStart struct {
	Name string
}

// 0x01 ->Server
type EncryptionResponse struct {
	// 0x01 ->Client
	SharedSecret []byte `mc:"len=short"`
	VerifyToken  []byte `mc:"len=short"`
}

// 0x00 ->Client
type LoginDisconnect struct {
	Reason string
}

// 0x02 ->Client
type LoginSuccess struct {
	// 0x02 ->Client
	UUID     string
	Username string
}

func (*LoginSuccess) StateUpdate() CxnState { return StatePlay }

// 0x01 ->Client
type EncryptionRequest struct {
	ServerId    string
	PublicKey   []byte `mc:"len=short"`
	VerifyToken []byte `mc:"len=short"`
}
