package webserver

import (
	"github.com/gorilla/pat"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	DOMAIN_NAME = "https://localhost:8080"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var router *pat.Router

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://localhost:8080"+r.RequestURI, http.StatusMovedPermanently)
}

type Receiver struct {
	err error
}

func newReceiver() *Receiver {
	return &Receiver{
		err: nil,
	}
}

func (r *Receiver) receive(ws *websocket.Conn, v interface{}) bool {
	r.err = ws.ReadJSON(v)
	return r.err == nil
}

// serveWs handles websocket requests from the peer.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	session := validateSessionAndLogInIfNecessary(w, r)
	if session == nil {
		return
	}
	log.Println("opening websocket")
	user, err := getUserFromSession(session)
	if err != nil {
		http.Error(w, "unable to retrieve user info", 500)
		return
	}
	userName := user.Name

	ws, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	receiver := newReceiver()
	chat.join <- ws
	var message Message
	for receiver.receive(ws, &message) {
		message.Username = userName
		message.Time = time.Now()
		chat.incoming <- &message
	}
	if receiver.err != nil {
		log.Println(receiver.err)
	}

	chat.leave <- ws
}

func Init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	chat = NewChat()
	router = pat.New()
	router.Get("/ws", wsHandler)

	//This has to be the last thing called with the router, because it sets
	//the handler for the website root.
	initAuth(router)
	http.Handle("/", router)
	http.Handle("/app/", http.StripPrefix("/app/",
		http.FileServer(http.Dir("app/"))))

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))
}
