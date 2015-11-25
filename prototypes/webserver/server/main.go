package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/gplus"
	"golang.org/x/oauth2/google"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	SESSION_NAME                     string = "user_session"
	SESSION_KEY_USERNAME                    = "username"
	GOOGLE_CLIENT_SECRET_FILE_PATH          = "../../../.gplus_client_secret.json"
	FACEBOOK_CLIENT_SECRET_FILE_PATH        = "../../../.facebook_client_secret.json"
	AUTH_CALLBACK_URL                       = "https://localhost:8080/oauth2callback"
	ONE_TIME_STATE_KEY                      = "one_time_state"
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

func endSession(s *sessions.Session, w http.ResponseWriter, r *http.Request) {
	s.Options = &sessions.Options{MaxAge: -1}
	s.Save(r, w)
}

func validateSessionAndLogInIfNecessary(
	w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := validateSession(w, r)
	if session == nil {
		if err != nil {
			log.Println(err.Error())
		}
		serveNewLogin(w, r)
	}

	return session
}

/**
 * return a session pointer. It is nil if the session could not be validated
 * (and thus the session is unauthorized). An error is also returned, if one
 * exists.
 */
func validateSession(
	w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	session, err := gothic.Store.Get(r, SESSION_NAME)
	log.Println("validating session...")

	if err != nil {
		log.Println("unable to get session.")
		return nil, err
	}

	if session.IsNew {
		endSession(session, w, r)
		return nil, nil
	}

	_, err = getUserFromSession(session)
	if err != nil {
		log.Println("unable to unmarshal user from session.")
		endSession(session, w, r)
		return nil, err
	}

	return session, nil
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth request")

	session, err := validateSession(w, r)
	if err != nil {
		log.Println(err)
	} else if session != nil {
		http.Redirect(w, r, "/app", http.StatusMovedPermanently)
		return
	}

	serveNewLogin(w, r)
}

func serveNewLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("serving new login.")
	t, err := template.New("login").Parse(indexTemplate)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	_, err = gothic.Store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, "unable to get session", 500)
		log.Println(err.Error())
	}

	t.Execute(w, nil)
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth callback")

	//TODO make use of more user attributes, besides name.
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	session, err := gothic.Store.Get(r, SESSION_NAME)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// log.Printf("number of values already in new session: %d.\n",
	// len(session.Values))

	err = putUserInSession(&user, session)
	if err != nil {
		http.Error(w, "unable to store user in session", 500)
		endSession(session, w, r)
		return
	}

	session.Save(r, w)
	http.Redirect(w, r, "/app", http.StatusMovedPermanently)
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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	googleJsonKey, err := ioutil.ReadFile(GOOGLE_CLIENT_SECRET_FILE_PATH)
	if err != nil {
		log.Fatalln("unable to read file ", GOOGLE_CLIENT_SECRET_FILE_PATH,
			":", err)
	}
	facebookJsonKey, err := ioutil.ReadFile(FACEBOOK_CLIENT_SECRET_FILE_PATH)
	if err != nil {
		log.Fatalln("unable to read file ", FACEBOOK_CLIENT_SECRET_FILE_PATH,
			":", err)
	}

	// do I need more scopes?
	// https://developers.google.com/+/domains/authentication/scopes
	googleConfig, err := google.ConfigFromJSON(googleJsonKey)
	if err != nil {
		log.Fatalln("unable to get google provider config:", err)
	}
	facebookConfig := &genericConfig{}
	err = json.Unmarshal(facebookJsonKey, facebookConfig)
	if err != nil {
		log.Fatalln("unable to get facebook provider config:", err)
	}

	//I need "profile", "email", scopes. gplus and facebook provide these by
	//default.
	goth.UseProviders(
		gplus.New(googleConfig.ClientID, googleConfig.ClientSecret,
			AUTH_CALLBACK_URL),
		facebook.New(facebookConfig.Client_id, facebookConfig.Client_secret, AUTH_CALLBACK_URL),
	)

	//Set the gothic store. It is also the store I will be using in general.
	initStore()

	chat = NewChat()
	router = pat.New()

	router.Get("/ws", wsHandler)
	router.Get("/oauth2callback", authCallbackHandler)
	router.Get("/auth/{provider}", gothic.BeginAuthHandler)
	router.Get("/", authHandler)
	// router.Add("GET", "/app", http.FileServer(http.Dir("app/")))
	// router.PathPrefix("/app").Handler(http.FileServer(http.Dir("app/")))
	http.Handle("/", router)
	http.Handle("/app/", http.StripPrefix("/app/",
		http.FileServer(http.Dir("app/"))))

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))
}

//TODO add more providers
var indexTemplate = `
<p><a href="/auth/gplus">Log in with Google</a></p>
<p><a href="/auth/facebook">Log in with Facebook</a></p>
`

type genericConfig struct {
	Client_id     string `json:"client_id"`
	Client_secret string `json:"client_secret"`
}