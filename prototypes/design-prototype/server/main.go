package main

import (
	"fmt"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8000", http.FileServer(http.Dir("app/")))
	if err != nil {
		fmt.Println(err)
	}
}
