package main

import (
	"flag"
	"fmt"
)

var (
	username = flag.String("u", "", "Minecraft user name")
	password = flag.String("p", "", "Minecraft password")
	server   = flag.String("s", "", "Minecraft server address")
)

func main() {
	flag.Parse()

	cfg, err := NewUserConfig(".mcbot-config")
	if err != nil {
		panic(err)
	}

	if *username != "" {
		cfg.SetValue("username", *username)
	}
	if *password != "" {
		cfg.SetSecret("password", *password)
	}

	if *server != "" {
		cfg.SetValue("server", *server)
	}

	if cfg.Value("server") == "" {
		cfg.SetValue("server", "localhost:25565")
	}

	if cfg.Value("username") == "" || cfg.Secret("password") == "" {
		fmt.Println("missing server/username/password. Specify once to first to have them saved in config")
		return
	}

	fmt.Println("Connecting", cfg.Value("server"))

	c, err := Connect(cfg)
	if err != nil {
		panic(err)
	}

	err = c.Login()
	if err != nil {
		panic(err)
	}
}
