package main

import (
	"log"
	// "../server"
	"../webserver"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//Start the backend before starting the webserver that relies upon it.
	//(William doesn't know how to get the back-end to work, so he's commented
	// it out.)
	// server.Init()
	//This is a blocking call. It just serves forever.
	webserver.Serve()

}
