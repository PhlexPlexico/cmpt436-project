package webserver

import (
	"encoding/json"
	"github.com/gorilla/pat"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	configFilePath = ".config.json"
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

//These will be marshaled directly from json
type config struct {
	Gplus          genericAuthConfig `json:"gplus"`
	Facebook       genericAuthConfig `json:"facebook"`
	Session_secret string            `json:"session_secret"`
	Website_url    string            `json:"website_url"`
	Https_portNum  string            `json:"https_portnum"`
	Http_portNum   string            `json:"http_portnum"`
}

func Serve() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln("unable to read file ", configFilePath,
			":", err)
	}

	conf := &config{}
	err = json.Unmarshal(configBytes, conf)
	if err != nil {
		log.Fatalln("unable to unmarshal config file:", err)
	}

	fm = NewFeedsManager()
	router = pat.New()
	router.Get("/ws", wsHandler)

	//This has to be the last thing called with the router, because it sets
	//the handler for the website root.
	initAuth(router, conf)
	http.Handle("/", router)
	http.Handle("/app/", http.StripPrefix("/app/",
		http.FileServer(http.Dir("app/"))))

	go log.Fatal(http.ListenAndServeTLS(conf.Https_portNum, "cert.crt", "key.key", nil))
	log.Fatal(http.ListenAndServe(conf.Http_portNum, http.HandlerFunc(redirectHandler)))
}
