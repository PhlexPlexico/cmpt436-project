package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	SESSION_DURATION_MINUTES int    = 30
	SESSION_NAME             string = "chat-session"
	HTTP_PARAM_USERNAME      string = "username"
	SESSION_KEY_USERNAME     string = HTTP_PARAM_USERNAME
)

var activeSessions = sessions.NewCookieStore([]byte("a-secret-key-i-guess"))
var chat *Chat

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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
		client.WriteJSON(message)
	}
	log.Println("broadcasted message")
}

type Message struct {
	Username string    `json:"username"`
	Time     time.Time `json:"time"`
	Message  string    `json:"message"`
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
	r.err = ws.ReadJSON(v)
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

	session, err := activeSessions.Get(r, SESSION_NAME)
	if err != nil {
		log.Println(err.Error())
		// http.Error(w, err.Error(), 500)
		// return
	}

	log.Printf("number of values already in new session: %d.\n", len(session.Values))
	userName := r.URL.Query().Get(HTTP_PARAM_USERNAME)
	if userName == "" {
		log.Println("nonexistent username in authentication request.")
		session.Options = &sessions.Options{MaxAge: -1}
		session.Save(r, w)
		return
	}

	session.Values[SESSION_KEY_USERNAME] = userName
	session.Save(r, w)
}

type WsHandler struct {
	err error
}

func NewWsHandler() *WsHandler {
	return &WsHandler{
		err: nil,
	}
}

// serveWs handles websocket requests from the peer.
func (wsh *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	existingSession, err := activeSessions.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//TODO inform the user of invalidity
	if existingSession.IsNew {
		log.Println("active session not found.")
		existingSession.Options = &sessions.Options{MaxAge: -1}
		existingSession.Save(r, w)
		return
	}

	existingSession.Save(r, w)

	ws, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Println(err)
		return
	}

	receiver := NewReceiver()
	chat.join <- ws
	var message Message
	for receiver.receive(ws, &message) {
		message.Username = existingSession.Values[SESSION_KEY_USERNAME].(string)
		message.Time = time.Now()
		chat.incoming <- &message
	}
	if receiver.err != nil {
		log.Println(receiver.err)
	}

	chat.leave <- ws
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	activeSessions.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * SESSION_DURATION_MINUTES,
		HttpOnly: true,
		Secure:   true,
	}

	chat = NewChat()

	r := mux.NewRouter()

	r.Handle("/ws", NewWsHandler())
	r.Handle("/auth", NewAuthHandler())
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))

}
