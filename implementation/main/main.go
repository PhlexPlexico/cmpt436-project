package main

import (
	"log"
	// "../db"
	"../server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//Start the backend before starting the server that relies upon it.
	//(William doesn't know how to get the back-end to work, so he's commented
	// it out.)
	// db.Init()
	//This is a blocking call. It just serves forever.
	server.Serve()
}
