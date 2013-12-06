package main

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

func (Handshake) Id(PktDisp) uint { return 0x00 }

// StateStatus
////////////////////////////////////////////////////////////////////////////////

type StatusRequest struct{}

// 0x00 ->Server
func (StatusRequest) Id(PktDisp) uint { return 0x00 }

type StatusResponse string

/*type StatusResponse struct {
	JsonData string
}*/
// 0x00 ->Client
func (StatusResponse) Id(PktDisp) uint { return 0x00 }

type StatusPing struct {
	// 0x01 ->Server ->Client
	Time int64
}

func (StatusPing) Id(PktDisp) uint { return 0x01 }

// StateLogin
////////////////////////////////////////////////////////////////////////////////

type LoginStart struct {
	// 0x00 ->Server
	Name string
}

func (LoginStart) Id(PktDisp) uint { return 0x00 }

type LoginDisconnect string

func (LoginDisconnect) Id(PktDisp) uint { return 0x00 }

type LoginSuccess struct {
	// 0x02 ->Client
	UUID     string
	Username string
}

func (LoginSuccess) Id(PktDisp) uint { return 0x02 }

type EncryptionRequest struct {
	// 0x01 ->Server
	ServerId    string
	PublicKey   []byte `mc:"len=int16"`
	VerifyToken []byte `mc:"len=int16"`
}

func (EncryptionRequest) Id(PktDisp) uint { return 0x01 }

type EncryptionResponse struct {
	// 0x01 ->Client
	SharedSecret []byte `mc:"len=int16"`
	VerifyToken  []byte `mc:"len=int16"`
}

func (EncryptionResponse) Id(PktDisp) uint { return 0x01 }

// StatePlay
////////////////////////////////////////////////////////////////////////////////

type KeepAlive struct {
	// 0x00 ->Server ->Client
	KeepAliveID int32
}

func (KeepAlive) Id(PktDisp) uint { return 0x00 }
