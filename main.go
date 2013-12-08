package main

import (
	"flag"
	"fmt"
	"github.com/tajtiattila/passwdprompt"
	"os"
)

var (
	server = flag.String("addr", "", "Minecraft server address")
)

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

	c, err := Connect(addr, NewConfigStore("auth", cfg))
	if err != nil {
		fail(err)
	}

	err = c.Login(UserPassworderFunc(func() (u, p string, err error) {
		return passwdprompt.GetUserPassword("Username: ", "Password: ")
	}))
	if err != nil {
		fail(err)
	}

	err = c.Run()
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Println(err)
	os.Exit(0)
}
