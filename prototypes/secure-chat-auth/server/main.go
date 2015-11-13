package main

import (
	"fmt"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
	"golang.org/x/oauth2/google"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	SESSION_DURATION_MINUTES       int    = 30
	SESSION_NAME                   string = "chat-session"
	HTTP_PARAM_USERNAME                   = "username"
	SESSION_KEY_USERNAME                  = HTTP_PARAM_USERNAME
	GOOGLE_CLIENT_SECRET_FILE_PATH        = "../../../.gplus_client_secret.json"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var router *pat.Router

var activeSessions = sessions.NewCookieStore([]byte("a-secret-key-i-guess"))

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

func AuthChoiceHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth request")

	t, err := template.New("foo").Parse(indexTemplate)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	t.Execute(w, nil)
}

func AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth callback")

	//TODO make use of more user attributes, besides name.
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	session, err := activeSessions.Get(r, SESSION_NAME)
	if err != nil {
		log.Println(err.Error())
		// http.Error(w, err.Error(), 500)
		// return
	}

	// t, err := template.New("foo").Parse(userTemplate)
	// if err != nil {
	// 	fmt.Fprintln(w, err)
	// 	return
	// }

	// t.Execute(w, user)

	log.Printf("number of values already in new session: %d.\n", len(session.Values))

	session.Values[SESSION_KEY_USERNAME] = user.Name
	session.Save(r, w)

	http.ServeFile(w, r, "app/")
}

// serveWs handles websocket requests from the peer.
func WsHandler(w http.ResponseWriter, r *http.Request) {
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

	jsonKey, err := ioutil.ReadFile(GOOGLE_CLIENT_SECRET_FILE_PATH)
	if err != nil {
		log.Println("unable to read file ", GOOGLE_CLIENT_SECRET_FILE_PATH)
		return
	}
	//TODO do I need scopes?
	// https://developers.google.com/+/domains/authentication/scopes
	googleConfig, err := google.ConfigFromJSON(jsonKey)
	goth.UseProviders(
		gplus.New(googleConfig.ClientID, googleConfig.ClientSecret,
			googleConfig.RedirectURL),
	)

	activeSessions.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * SESSION_DURATION_MINUTES,
		HttpOnly: true,
		Secure:   true,
	}

	chat = NewChat()
	router = pat.New()

	router.Get("/ws", WsHandler)
	router.Get("/app", AuthCallbackHandler)
	router.Get("/auth/{provider}", gothic.BeginAuthHandler)
	router.Get("/", AuthChoiceHandler)
	http.Handle("/", router)

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))
}

var indexTemplate = `
<p><a href="/auth/gplus">Log in with Google</a></p>
`

var userTemplate = `
<p>Name: {{.Name}}</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
`
