package main

import (
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

func wsHandler(ws *websocket.Conn) {
	ws.Close()
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://localhost:8080"+r.RequestURI, http.StatusMovedPermanently)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	r := mux.NewRouter()

	r.Handle("/ws", websocket.Handler(wsHandler))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))
}
