package main

import (
	"flag"
	"fmt"
	mcnet "github.com/tajtiattila/mctoy/net"
	proto "github.com/tajtiattila/mctoy/protocol"
	"github.com/tajtiattila/passwdprompt"
	"io"
	"os"
	"reflect"
	"sync"
	"time"
)

var (
	server = flag.String("addr", "", "Minecraft server address")
)

type DemoHandler struct {
	mtx        sync.Mutex
	responder  bool
	PlayerID   int32
	X, Y, Z    float64
	Yaw, Pitch float32
	OnGround   bool
	log        io.ReadWriter
}

func (h *DemoHandler) SendPosition(c *mcnet.Conn) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	if h.PlayerID == 0 {
		// not joined yet
		return nil
	}
	return c.Send(proto.ClientPlayerPositionAndLook{
		X:        h.X,
		Y:        h.Y,
		Z:        h.Z,
		Stance:   2.0,
		Yaw:      h.Yaw,
		Pitch:    h.Pitch,
		OnGround: h.OnGround,
	})
}

func (h *DemoHandler) HandlePacket(c *mcnet.Conn, pk interface{}) (err error) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	fmt.Fprintln(h.log, reflect.TypeOf(pk))
	switch p := pk.(type) {
	case *proto.KeepAlive:
		err = c.Send(p)
		return
	case *proto.JoinGame:
		h.PlayerID = p.EntityID
	case *proto.ServerPlayerPositionAndLook:
		h.X, h.Y, h.Z = p.X, p.Y, p.Z
		h.Yaw, h.Pitch = p.Yaw, p.Pitch
		h.OnGround = p.OnGround
	case *proto.MapChunkBulk:
		if !h.responder {
			h.responder = true
			go func() {
				for _ = range time.Tick(time.Second / 20) {
					err := h.SendPosition(c)
					fmt.Fprintln(h.log, "-> SendPosition:", err)
				}
			}()
		}
	}
	return
}

func main() {
	flag.Parse()

	cfg, err := NewUserConfig(".mcbot-config")
	if err != nil {
		panic(err)
	}

	if *server != "" {
		cfg.SetValue("server", *server)
	}

	if cfg.Value("server") == "" {
		cfg.SetValue("server", "localhost:25565")
	}

	addr := cfg.Value("server")
	fmt.Println("Connecting", addr)

	c, err := mcnet.Connect(addr)
	if err != nil {
		fail(err)
	}

	var a mcnet.Auth
	a = mcnet.NewYggAuth(
		NewConfigStore("auth", cfg),
		mcnet.UserPassworderFunc(func() (u, p string, err error) {
			return passwdprompt.GetUserPassword("Username: ", "Password: ")
		}),
	)
	a = mcnet.NewNoAuth("Sándorvagyok")
	err = c.Login(a)
	if err != nil {
		fail(err)
	}

	h := &DemoHandler{log: NewRoundBuf()}
	err = c.Run(h)
	//io.Copy(os.Stdout, h.log)

	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Println(err)
	os.Exit(0)
}
