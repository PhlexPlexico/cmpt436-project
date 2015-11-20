package main

import (
	//"../logic"
	"../server"
)

func main() {
	var err error
	server.ThisPanic(err)
	server.ConnectToDB()
}
