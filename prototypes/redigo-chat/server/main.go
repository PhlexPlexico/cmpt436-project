package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var chat *Chat

type Chat struct {
	join     chan *websocket.Conn
	incoming chan *Message
	clients  []*websocket.Conn
}

func NewChat() *Chat {
	chat := &Chat{
		join:     make(chan *websocket.Conn),
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
			case message := <-chat.incoming:
				chat.Broadcast(message)
			}
		}
	}()
}

func (chat *Chat) Join(client *websocket.Conn) {
	chat.clients = append(chat.clients, client)
	go func() {
		for {
			var message Message
			err := client.ReadJSON(&message)
			message.Time = time.Now()
			if err != nil {
				log.Println(err)
				break
			}
			chat.incoming <- &message
			log.Println("decoded message")
		}
	}()
}

func (chat *Chat) Broadcast(message *Message) {
	for _, client := range chat.clients {
		client.WriteJSON(message)
	}
}

type Message struct {
	Username string    `json:"username"`
	Time     time.Time `json:"time"`
	Message  string    `json:"message"`
}

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	chat.join <- ws
}

func main() {
	chat = NewChat()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	r := mux.NewRouter()
	r.HandleFunc("/ws", serveWs)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
