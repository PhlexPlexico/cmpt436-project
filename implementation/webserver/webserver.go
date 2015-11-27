package webserver

import (
	"encoding/json"
	"github.com/gorilla/pat"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const (
	DOMAIN_NAME = "https://localhost:8080"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//We shouldn't need a CheckOrigin function, because at this point the session is
	//already validated.
}

var router *pat.Router

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://localhost:8080"+r.RequestURI, http.StatusMovedPermanently)
}

type receiver struct {
	err error
}

func newReceiver() *receiver {
	return &receiver{
		err: nil,
	}
}

func (r *receiver) receive(ws *websocket.Conn, v interface{}) bool {
	r.err = ws.ReadJSON(v)
	return r.err == nil
}

// handles websocket requests from the client.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	session := validateSessionAndLogInIfNecessary(w, r)
	if session == nil {
		return
	}

	log.Println("opening websocket")

	// may not need user name, but if we do, we can get it like this.
	// user, err := getUserFromSession(session)
	// if err != nil {
	// 	http.Error(w, "unable to retrieve user info",
	// 		http.StatusInternalServerError)
	// 	return
	// }
	// userName := user.Name

	ws, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	serveWs(ws)
	// receiver := newReceiver()
	// fm.join <- ws
	// var message Message
	// for receiver.receive(ws, &message) {
	// 	message.Username = userName
	// 	message.Time = time.Now()
	// 	chat.incoming <- &message
	// }
	// if receiver.err != nil {
	// 	log.Println(receiver.err)
	// }

	// chat.leave <- ws
}

func writeJson(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Fatalln(err)
	}
}

func Serve() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fm = NewFeedsManager()
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
