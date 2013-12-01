package main

type PacketKind int

type State uint

const (
	StateHandshake State = 0
	StateStatus    State = 1
	StateLogin     State = 2
	StatePlay      State = 3
)

type Direction int

const (
	ToServer Direction = 1
	ToClient Direction = 2
	ToAny    Direction = 3
)

type PacketIdent struct {
	St  State
	Id  int
	Dir Direction
}

type Packet interface {
	Id() uint
}

// StateHandshake
////////////////////////////////////////////////////////////////////////////////

type Handshake struct {
	// 0x00 ->Server
	ProtocolVersion uint
	ServerAddress   string
	ServerPort      uint16
	NextState       State
}

func (Handshake) Id() uint { return 0x00 }

// StateStatus
////////////////////////////////////////////////////////////////////////////////

type StatusRequest struct{}

// 0x00 ->Server
func (StatusRequest) Id() uint { return 0x00 }

type StatusResponse string

/*type StatusResponse struct {
	JsonData string
}*/
// 0x00 ->Client
func (StatusResponse) Id() uint { return 0x00 }

type Ping struct {
	// 0x01 ->Server ->Client
	Time int64
}

func (Ping) Id() uint { return 0x01 }

// StateLogin
////////////////////////////////////////////////////////////////////////////////

type KeepAlive struct {
	// 0x00 ->Server ->Client
	KeepAliveID int32
}

func (KeepAlive) Id() uint { return 0x00 }

type LoginStart struct {
	// 0x00 ->Server
	Name string
}

func (LoginStart) Id() uint { return 0x00 }

type Disconnect struct {
	// 0x00 ->Client
}

func (Disconnect) Id() uint { return 0x00 }

type LoginSuccess struct {
	// 0x02 ->Client
	UUID     string
	Username string
}

func (LoginSuccess) Id() uint { return 0x02 }

type EncryptionRequest struct {
	// 0x01 ->Server
	ServerId    string
	PublicKey   []byte `mc:"len=int16"`
	VerifyToken []byte `mc:"len=int16"`
}

func (EncryptionRequest) Id() uint { return 0x01 }

type EncryptionResponse struct {
	// 0x01 ->Client
	SharedSecret []byte `mc:"len=int16"`
	VerifyToken  []byte `mc:"len=int16"`
}

func (EncryptionResponse) Id() uint { return 0x01 }
