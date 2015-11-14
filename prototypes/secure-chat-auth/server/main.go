package main

import (
	"fmt"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/markbates/goth"
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
	SESSION_NAME                   string = "user_session"
	SESSION_KEY_USERNAME                  = "username"
	GOOGLE_CLIENT_SECRET_FILE_PATH        = "../../../.gplus_client_secret.json"
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

func NewReceiver() *Receiver {
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

/**
 * return true if a pre-existing user was attached, false otherwise.
 * if false, an associated error may be returned (unless there simply was no
 * pre-existing user.
 */
func tryToAttachPreexistingUserToSession(
	w http.ResponseWriter, r *http.Request) (bool, error) {
	session, err := Store.Get(r, SESSION_NAME)
	if err != nil {
		log.Println("unable to get session.")
		return false, err
	}

	if session.IsNew {
		return false, nil
	}

	provider, err := getProviderFromSession(session)
	if err != nil {
		log.Println("unable to get provider")
		return false, err
	}

	sess, err := provider.UnmarshalSession(session.Values[GOTH_SESS_KEY].(string))
	if err != nil {
		log.Println("unable to unmarshal sess from session")
		endSession(session, w, r)
		return false, err
	}

	user, err := provider.FetchUser(sess)
	//TODO generalize 'user.RawData["error"] != nil'.
	//Works for gplus, but unlikely to work for all providers.
	if err != nil || user.RawData["error"] != nil {
		log.Println("unable to fetch user with ", provider.Name())
		endSession(session, w, r)
		return false, err
	}

	err = putUserInSession(&user, session)
	if err != nil {
		log.Println("unable to store user in session.")
		endSession(session, w, r)
		return false, err
	}

	session.Save(r, w)
	return true, nil
}

func AuthChoiceHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth request")

	wasUserFound, err := tryToAttachPreexistingUserToSession(w, r)
	if err != nil {
		log.Println(err)
	}
	if wasUserFound {
		http.Redirect(w, r, "/realapp", http.StatusMovedPermanently)
		return
	}

	log.Println("serving new login.")
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
	user, err := CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	session, err := Store.Get(r, SESSION_NAME)

	if err != nil {
		log.Println(err.Error())
		//Apparently this err does not mean no session was created, so I don't need
		//to return.
		// http.Error(w, err.Error(), 500)
		// return
	}

	log.Printf("number of values already in new session: %d.\n", len(session.Values))

	err = putUserInSession(&user, session)
	if err != nil {
		http.Error(w, "unable to store user in session", 500)
		endSession(session, w, r)
		return
	}

	session.Save(r, w)
	log.Println("163")
	// http.ServeFile(w, r, "app/")
	http.Redirect(w, r, "/realapp", http.StatusMovedPermanently)
}

// serveWs handles websocket requests from the peer.
func WsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println("175")
	//TODO inform the user of invalidity
	if session.IsNew {
		http.Error(w, "invalid action: user is not logged in.", 401)
		endSession(session, w, r)
		return
	}

	//TODO try commenting this out, see if it makes a difference (it shouldn't)
	// session.Save(r, w)

	user, err := getUserFromSession(session)
	if err != nil {
		log.Println("188")
		log.Println(err.Error())
		http.Error(w, "unable to retrieve user from session", 500)
		endSession(session, w, r)
		return
	}
	log.Println("192")
	userName := user.Name

	ws, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	receiver := NewReceiver()
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

	jsonKey, err := ioutil.ReadFile(GOOGLE_CLIENT_SECRET_FILE_PATH)
	if err != nil {
		log.Println("unable to read file ", GOOGLE_CLIENT_SECRET_FILE_PATH)
		return
	}
	// do I need more scopes?
	// https://developers.google.com/+/domains/authentication/scopes
	googleConfig, err := google.ConfigFromJSON(jsonKey)

	//gives me "profile", "email", "openid" scopes by default.
	goth.UseProviders(
		gplus.New(googleConfig.ClientID, googleConfig.ClientSecret,
			googleConfig.RedirectURL),
	)

	chat = NewChat()
	router = pat.New()

	router.Get("/ws", WsHandler)
	router.Get("/app", AuthCallbackHandler)
	router.Get("/auth/{provider}", BeginAuthHandler)
	router.Get("/", AuthChoiceHandler)
	// router.Add("GET", "/realapp", http.FileServer(http.Dir("app/")))
	// router.PathPrefix("/realapp").Handler(http.FileServer(http.Dir("app/")))
	http.Handle("/", router)
	http.Handle("/realapp/", http.StripPrefix("/realapp/",
		http.FileServer(http.Dir("app/"))))

	go http.ListenAndServeTLS(":8080", "cert.crt", "key.key", nil)
	http.ListenAndServe(":8000", http.HandlerFunc(redirectHandler))
}

//TODO add more providers
var indexTemplate = `
<p><a href="/auth/gplus">Log in with Google</a></p>
`
