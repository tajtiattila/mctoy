package main

import (
	"flag"
	"fmt"
	mcnet "github.com/tajtiattila/mctoy/net"
	"github.com/tajtiattila/passwdprompt"
	"os"
)

var (
	server = flag.String("addr", "", "Minecraft server address")
)

type playHandler struct{}

func (*playHandler) HandlePacket(
	c *mcnet.Conn,
	pkid uint,
	pkname string,
	pk mcnet.Packet,
) (err error) {
	fmt.Printf("%02x %s\n", pkid, pkname)
	if k, ok := pk.(*mcnet.KeepAlive); ok {
		err = c.Send(k)
		return
	}
	//fmt.Printf("%#v\n\n", pk)
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

	c, err := mcnet.Connect(addr, NewConfigStore("auth", cfg))
	if err != nil {
		fail(err)
	}

	err = c.Login(mcnet.UserPassworderFunc(func() (u, p string, err error) {
		return passwdprompt.GetUserPassword("Username: ", "Password: ")
	}))
	if err != nil {
		fail(err)
	}

	err = c.Run(&playHandler{})
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Println(err)
	os.Exit(0)
}
