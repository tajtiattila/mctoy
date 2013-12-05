package main

import (
	"flag"
	"fmt"
	"github.com/tajtiattila/passwdprompt"
)

var (
	server = flag.String("s", "", "Minecraft server address")
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

	fmt.Println("Connecting", cfg.Value("server"))

	c, err := Connect(cfg)
	if err != nil {
		panic(err)
	}

	err = c.Login(UserPassworderFunc(func() (u, p string, err error) {
		return passwdprompt.GetUserPassword("Username: ", "Password: ")
	}))
	if err != nil {
		panic(err)
	}
}
