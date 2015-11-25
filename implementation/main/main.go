package main

import (
	"../server"
	"../webserver"
)

func main() {
	webserver.Init()
	server.Init()
}
