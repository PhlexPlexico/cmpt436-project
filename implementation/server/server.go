package server

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

func redirectHandler(conf *config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("https://" + conf.WebsiteUrl + conf.HttpsPortNum + r.RequestURI)
		http.Redirect(w, r, "https://"+conf.WebsiteUrl+conf.HttpsPortNum+r.RequestURI,
			http.StatusMovedPermanently)
	}
}

// handles websocket requests from the client.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	user := validateUserAndLogInIfNecessary(w, r)
	if user == nil {
		return
	}

	log.Println("opening websocket")

	ws, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	serveWs(ws, user.Id)
}

//These will be marshaled directly from json
type config struct {
	Gplus          genericAuthConfig `json:"gplus"`
	Facebook       genericAuthConfig `json:"facebook"`
	Session_secret string            `json:"session_secret"`
	WebsiteUrl     string            `json:"website_url"`
	HttpsPortNum   string            `json:"https_portnum"`
	HttpPortNum    string            `json:"http_portnum"`
	RestPortNum    string            `json:"rest_portnum"`
}

func Serve() {
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

	//Serve all the rest api calls.
	serveRestApi(conf)

	//This has to be the last thing called with the router, because it sets
	//the handler for the website root.
	initAuth(router, conf)
	http.Handle("/", router)
	//This static final can only be reached via explicit redirect: typing it into
	//the address bar just makes the router handle it.
	http.HandleFunc("/app/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[4:] == "/" {
			appHandler(w, r)
		} else {
			http.ServeFile(w, r, r.URL.Path[1:])
		}
	})

	log.Print("https://" + conf.WebsiteUrl + conf.HttpsPortNum)

	go func() {
		log.Fatal(http.ListenAndServeTLS(conf.HttpsPortNum,
			"cert.crt", "key.key", nil))
	}()

	log.Fatal(http.ListenAndServe(conf.HttpPortNum,
		http.HandlerFunc(redirectHandler(conf))))
}
