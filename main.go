package main

import (
	"fmt"
)

func main() {
	//ss, err := NewServerStatus("shuttle-xpc.local", 25565)
	//fmt.Println(ss, err)

	u, p := Config.Get("username"), Config.GetSecret("password")

	c, err := Connect("shuttle-xpc.local", 25565)
	fmt.Println(err)

	err = c.Login(u, p)
	fmt.Println(err)
}
