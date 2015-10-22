package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	SESSSION_ID_SIZE         int = 32
	SESSION_DURATION_MINUTES     = 30
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

func wsHandler(ws *websocket.Conn) {
	receiver := NewReceiver()
	var session Session
	receiver.receive(ws, &session)
	if receiver.err != nil {
		log.Println(receiver.err)
		return
	}

	existingSession, ok := getActiveSession(session.Username)

	//TODO inform the user of invalidity
	if !ok {
		log.Println("active session not found.")
		return
	}
	if existingSession.SessionId != session.SessionId {
		log.Println("invalid sessionId")
		return
	}
	if existingSession.ExpiryTime.Before(time.Now()) {
		log.Println("expired session")
		removeActiveSession(session.Username)
		return
	}

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
	query := r.URL.Query()
	username := query.Get("username")
	password := query.Get("password")
	//TODO actually authorize something

	//TODO account for the case where session is nil
	session := NewSession(username, password)
	addActiveSession(session)

	log.Println(session)
	var sessionJsonBytes []byte
	sessionJsonBytes, a.err = json.Marshal(session)
	if a.err != nil {
		log.Panic("session cannot be stored as a json object")
	}

	log.Println(string(sessionJsonBytes))
	w.Write(sessionJsonBytes)
}

type Session struct {
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	SessionId  string    `json:"sessionId"`
	ExpiryTime time.Time `json:"expiryTime"`
}

func NewSession(username, password string) *Session {
	sessionIdBytes := make([]byte, SESSSION_ID_SIZE)
	_, err := rand.Read(sessionIdBytes)
	if err != nil {
		log.Println("could not generate session id")
		return nil
		//TODO handle this properly
	}

	sessionId := base64.URLEncoding.EncodeToString(sessionIdBytes)
	expiryTime := time.Now().Add(time.Duration(SESSION_DURATION_MINUTES) * time.Minute)

	return &Session{
		Username:   username,
		Password:   password,
		SessionId:  sessionId,
		ExpiryTime: expiryTime,
	}
}

func (s *Session) String() string {
	// return fmt.Sprintf("session: {username: %s; password: "+
	// 	"%s; sessionId: %s; expiryTime: %s",
	// 	s.Username, s.Password, s.SessionId, s.ExpiryTime.String())
	json, err := json.Marshal(s)
	if err != nil {
		log.Panic("JSON can't marshal session.")
	}
	return string(json)
}

func getActiveSession(username string) (session *Session, ok bool) {
	lockActiveSessions.RLock()
	session, ok = activeSessions[username]
	lockActiveSessions.RUnlock()
	return
}

func addActiveSession(session *Session) {
	//TODO deal with duplicate usernames, don't just return true
	lockActiveSessions.Lock()
	activeSessions[session.Username] = session
	lockActiveSessions.Unlock()
	return
}

func removeActiveSession(username string) (ok bool) {
	lockActiveSessions.Lock()
	_, ok = activeSessions[username]
	if ok {
		delete(activeSessions, username)
	}
	lockActiveSessions.Unlock()
	return
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	activeSessions = make(map[string]*Session)

	chat = NewChat()

	r := mux.NewRouter()

	r.Handle("/ws", websocket.Handler(wsHandler))
	r.Handle("/auth", NewAuthHandler())
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("app/")))

	http.Handle("/", r)

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))

}
