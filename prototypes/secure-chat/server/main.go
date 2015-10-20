package main

import (
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

var chat *Chat

type Chat struct {
	join     chan *websocket.Conn
	leave    chan *websocket.Conn
	incoming chan *Message
	clients  []*websocket.Conn
}

func NewChat() *Chat {
	chat := &Chat{
		join:     make(chan *websocket.Conn),
		leave:    make(chan *websocket.Conn),
		incoming: make(chan *Message),
		clients:  make([]*websocket.Conn, 0),
	}
	chat.Listen()
	return chat
}

func (chat *Chat) Listen() {
	go func() {
		for {
			select {
			case client := <-chat.join:
				chat.Join(client)
			case client := <-chat.leave:
				chat.Leave(client)
			case message := <-chat.incoming:
				chat.Broadcast(message)
			}
		}
	}()
}

func (chat *Chat) Join(client *websocket.Conn) {
	chat.clients = append(chat.clients, client)
	log.Println("client joined")
}

func (chat *Chat) Leave(client *websocket.Conn) {
	chat.clients = append(chat.clients, client)
	for i, otherClient := range chat.clients {
		if client == otherClient {
			chat.clients = append(chat.clients[:i], chat.clients[i+1:]...)
			break
		}
	}
	log.Println("client left")
}

func (chat *Chat) Broadcast(message *Message) {
	for _, client := range chat.clients {
		websocket.JSON.Send(client, message)
	}
	log.Println("broadcasted message")
}

type Message struct {
	Username string    `json:"username"`
	Time     time.Time `json:"time"`
	Message  string    `json:"message"`
}

func wsHandler(ws *websocket.Conn) {
	chat.join <- ws
	for {
		var message Message
		err := websocket.JSON.Receive(ws, &message)
		if err != nil {
			log.Println(err)
			break
		}
		message.Time = time.Now()
		chat.incoming <- &message
	}
	chat.leave <- ws
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://localhost:8080"+r.RequestURI, http.StatusMovedPermanently)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	chat = NewChat()

	r := mux.NewRouter()

	r.Handle("/ws", websocket.Handler(wsHandler))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))
}
