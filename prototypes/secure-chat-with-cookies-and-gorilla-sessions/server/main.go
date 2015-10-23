package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	SESSION_DURATION_MINUTES int = 30
)

var activeSessions = sessions.NewCookieStore([]byte("a-secret-key-i-guess"))
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
	existingSession, err := activeSessions.Get(r, session.Username)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//TODO inform the user of invalidity
	if existingSession.IsNew() {
		log.Println("active session not found.")
		existingSession.Options = &activeSessions.Options{MaxAge: -1}
		existingSession.Save(r, w)
		return
	}

	receiver := NewReceiver()
	chat.join <- ws
	var message Message
	for receiver.receive(ws, &message) {
		message.Username = session.Username
		message.Time = time.Now()
		chat.incoming <- &message
	}
	if receiver.err != nil {
		log.Println(receiver.err)
	}

	chat.leave <- ws
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://localhost:8080"+r.RequestURI, http.StatusMovedPermanently)
}

type Receiver struct {
	err error
}

func NewReceiver() *Receiver {
	return &Receiver{
		err: nil,
	}
}

func (r *Receiver) receive(ws *websocket.Conn, v interface{}) bool {
	r.err = websocket.JSON.Receive(ws, v)
	return r.err == nil
}

type AuthHandler struct {
	err error
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		err: nil,
	}
}

func (a *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("serving http request")
	// if direct access is needed, use this syntax.
	// r.URL.Query().Get("username")
	session, err := activeSessions.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Println("number of values already in new session: " + len(session.Values))
	session.Save(r, w)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	activeSessions.Options{
		Path:     "/",
		MaxAge:   60 * SESSION_DURATION_MINUTES,
		HttpOnly: true,
	}

	chat = NewChat()

	r := mux.NewRouter()

	r.Handle("/ws", websocket.Handler(wsHandler))
	r.Handle("/auth", NewAuthHandler())
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))

}
