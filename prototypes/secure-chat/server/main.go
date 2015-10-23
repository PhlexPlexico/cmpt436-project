package main

import (
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

const (
	REDIS_ADDRESS = "localhost:6379"
)

func wsHandler(ws *websocket.Conn) {
	client := NewClient(ws)
	chat.join <- client
	for {
		var message Message
		err := websocket.JSON.Receive(ws, &message)
		if err != nil {
			log.Println(err)
			break
		}
		message.Time = time.Now().String()
		chat.incoming <- &message
	}
	chat.leave <- client
}

// redirects http requests to https
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://localhost:8080"+r.RequestURI, http.StatusMovedPermanently)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	pool = NewPool(REDIS_ADDRESS)
	chat = NewChat()

	r := mux.NewRouter()

	r.Handle("/ws", websocket.Handler(wsHandler))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)

	go func() {
		err := http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
		if err != nil {
			log.Println(err)
		}
	}()
	err := http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))
	if err != nil {
		log.Println(err)
	}
}
