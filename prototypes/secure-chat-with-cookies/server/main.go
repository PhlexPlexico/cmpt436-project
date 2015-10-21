package main

import (
	"encoding/json"
	"encoding/base64"
	"crypto/rand"
	"sync"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

const (
	SESSSION_ID_SIZE int = 32
	SESSION_DURATION_MINUTES = 30
)

var activeSessions map[string]*Session
var lockActiveSessions sync.RWMutex
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

type Session struct {
	username string	`json:"username"`
	sessionId string `json:"sessionId"`
	expiryTime time.Time `json:"expiryTime"`
}

func NewSession(username string) *Session {
	sessionIdBytes := make([]byte, SESSSION_ID_SIZE)
	n, err := rand.Read(sessionIdBytes)
	if err != nil {
		log.Println("could not generate session id")
		return nil
		//TODO handle this properly
	}

	sessionId := string(sessionIdBytes[:n])
	expiryTime := time.Now().Add(time.Duration(SESSION_DURATION_MINUTES) * time.Minute)
	
	return &Session {
		username: username,
		sessionId: sessionId,
		expiryTime: expiryTime,
	}
}
func authHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	username := query.Get("username")
	password := query.Get("password")
	//TODO actually authorize something

	session := NewSession()
	addActiveSession(session)

	//TODO account for the case where session is nil
	w.Write(json.Marshal(session))
}

func wsHandler(ws *websocket.Conn) {
	receiver := NewReceiver()
	var session Session
	receiver.receive(ws, &session)
	if receiver.handleErr() {
		return
	}

	existingSession, ok := getActiveSession(session.username)

	//TODO inform the user of invalidity
	if !ok || existingSession.sessionId != session.sessionId {
		log.Println("invalid sessionId")
		return
	} else if existingSession.expiryTime.Before(time.Now()) {
		log.Println("expired session")
		removeActiveSession(session.username)
		return
	}

	chat.join <- ws
	var message Message
	for receiver.receive(ws, &message) {
		message.Username = session.
		message.Time = time.Now()
		chat.incoming <- &message
	}
	receiver.handleErr()

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

func (r *Receiver) handleErr() bool {
	if r.err != nil {
		log.Println(err)
	}

	return r.err != nil
}

func getActiveSession(username string) (session *Session, ok bool) {
	lockActiveSessions.RLock()
	session, ok = activeSessions[username]
	lockActiveSessions.RUnlock()
	return
}

func addActiveSession(session *Session) (ok bool) {
	//TODO deal with duplicate usernames
	lockActiveSessions.Lock()
	activeSessions[session.username] = session
	lockActiveSessions.Unlock()
	ok = true
}

func removeActiveSession(username string) (ok bool) {
	lockActiveSessions.Lock()
	_, ok = activeSessions[username]
	if ok {
		delete(activeSessions, username)
	} 
	lockActiveSessions.Unlock()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	activeSessions = make(map[string]*Session)

	chat = NewChat()

	r := mux.NewRouter()

	r.Handle("/ws", websocket.Handler(wsHandler))
	r.Handle("/auth", http.HandleFunc(authHandler))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))

}