package net

type PacketKind int

// StateHandshake
////////////////////////////////////////////////////////////////////////////////

type Handshake struct {
	// 0x00 ->Server
	ProtocolVersion uint
	ServerAddress   string
	ServerPort      uint16
	NextState       uint
}

func (Handshake) Id() (uint, uint) { return PktInvalid, 0x00 }

// StateStatus
////////////////////////////////////////////////////////////////////////////////

type StatusRequest struct{}

// 0x00 ->Server
func (StatusRequest) Id() (uint, uint) { return PktInvalid, 0x00 }

type StatusResponse struct {
	JSON string
}

// 0x00 ->Client
func (StatusResponse) Id() (uint, uint) { return 0x00, PktInvalid }

type StatusPing struct {
	// 0x01 ->Server ->Client
	Time int64
}

func (StatusPing) Id() (uint, uint) { return 0x01, 0x01 }

// StateLogin
////////////////////////////////////////////////////////////////////////////////

type LoginStart struct {
	// 0x00 ->Server
	Name string
}

func (LoginStart) Id() (uint, uint) { return PktInvalid, 0x00 }

type LoginDisconnect struct {
	Reason string
}

func (LoginDisconnect) Id() (uint, uint) { return 0x00, PktInvalid }

type LoginSuccess struct {
	// 0x02 ->Client
	UUID     string
	Username string
}

func (LoginSuccess) Id() (uint, uint) { return 0x02, PktInvalid }

type EncryptionRequest struct {
	// 0x01 ->Server
	ServerId    string
	PublicKey   []byte `mc:"len=int16"`
	VerifyToken []byte `mc:"len=int16"`
}

func (EncryptionRequest) Id() (uint, uint) { return 0x01, PktInvalid }

type EncryptionResponse struct {
	// 0x01 ->Client
	SharedSecret []byte `mc:"len=int16"`
	VerifyToken  []byte `mc:"len=int16"`
}

func (EncryptionResponse) Id() (uint, uint) { return PktInvalid, 0x01 }
