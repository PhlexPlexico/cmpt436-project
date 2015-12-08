package main

import (
	"../db"
	"../server"
	"fmt"
	"log"
	"os"
)

func main() {

	if os.Args[1] == "test" {
		fmt.Println(len(os.Args), os.Args[len(os.Args)-1])

		db.Test()
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		//Start the backend before starting the server that relies upon it.
		//(William doesn't know how to get the back-end to work, so he's commented
		// it out.)
		db.Init()
		//This is a blocking call. It just serves forever.
		server.Serve()

	}

}
