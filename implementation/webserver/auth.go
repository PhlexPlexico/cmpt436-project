/*
 * Uses goth/gothic for authentication, and also makes use of the session that
 * gothic uses (so that there are not two sessions being used.)
 */
package webserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/gplus"
	"html/template"
	"log"
	"net/http"
)

const (
	sessionDurationMinutes   int    = 30
	userKey                  string = "goth_user"
	authCallbackRelativePath        = "/oauth2callback"
)

func marshalUser(user *goth.User) (string, error) {
	b, err := json.Marshal(user)
	return string(b), err
}

func unmarshalUser(data string) (*goth.User, error) {
	user := &goth.User{}
	err := json.Unmarshal([]byte(data), user)
	return user, err
}

func getUserFromSession(s *sessions.Session) (*goth.User, error) {
	val := s.Values[userKey]
	if val == nil {
		return nil, errors.New("user not stored in session")
	}
	userString := val.(string)
	return unmarshalUser(userString)
}

/*
 * Does not save the session.
 */
func putUserInSession(user *goth.User, s *sessions.Session) error {
	userString, err := marshalUser(user)
	if err != nil {
		return err
	}
	s.Values[userKey] = userString
	return nil
}

/**
 * Validate the session for this request. If it is invalid, serve a new login.
 * @return the session, if valid, or nil if serving a new login
 */
func validateSessionAndLogInIfNecessary(
	w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := validateSession(w, r)
	if session == nil {
		if err != nil {
			log.Println(err.Error())
		}
		serveNewLogin(w, r)
		return nil
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
	session, err := getSession(r)
	log.Println("validating session...")

	if err != nil {
		log.Println("unable to get session.")
		return nil, err
	}

	if session.IsNew {
		return nil, nil
	}

	_, err = getUserFromSession(session)
	if err != nil {
		log.Println("unable to unmarshal user from session.")
		return nil, err
	}

	return session, nil
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth request")

	session := validateSessionAndLogInIfNecessary(w, r)
	if session != nil {
		http.Redirect(w, r, "/app", http.StatusMovedPermanently)
	}
}

func serveNewLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("serving new login.")
	t, err := template.New("login").Parse(indexTemplate)
	if err != nil {
		fmt.Fprintln(w, err)
		return
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

	session, err := getSession(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = putUserInSession(&user, session)
	if err != nil {
		http.Error(w, "unable to store user in session",
			http.StatusInternalServerError)
		endSession(session, w, r)
		return
	}

	session.Save(r, w)
	http.Redirect(w, r, "/app", http.StatusMovedPermanently)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving logout")
	session, err := getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	endSession(session, w, r)
	serveNewLogin(w, r)
}

// func handleError(err error, s *sessions.Session,
// 	w http.ResponseWriter, r *http.Request) {
// 	if s == nil {
// 		s, err2 = getSession(r)
// 		if err2 != nil {
// 			http.Error(w, err2.Error(), http.StatusInternalServerError)
// 			log.Println(err2.Error())
// 			return
// 		}
// 	}

// 	endSession(s, w, r)
// }

func endSession(s *sessions.Session, w http.ResponseWriter, r *http.Request) {
	s.Options = &sessions.Options{MaxAge: -1}
	s.Save(r, w)
}

func getSession(r *http.Request) (*sessions.Session, error) {
	return gothic.Store.Get(r, gothic.SessionName)
}

//TODO add more providers
var indexTemplate = `
<p><a href="/auth/gplus">Log in with Google</a></p>
<p><a href="/auth/facebook">Log in with Facebook</a></p>
`

type genericAuthConfig struct {
	Client_id     string `json:"client_id"`
	Client_secret string `json:"client_secret"`
}

func initAuth(router *pat.Router, conf *config) {
	//get all the providers set up.
	//I need "profile", "email", scopes. gplus and facebook provide these by
	//default.
	AUTH_CALLBACK_PATH := "https://" + conf.Website_url + conf.Https_portNum + authCallbackRelativePath
	goth.UseProviders(
		gplus.New(conf.Gplus.Client_id, conf.Gplus.Client_secret,
			AUTH_CALLBACK_PATH+"/gplus"),
		facebook.New(conf.Facebook.Client_id, conf.Facebook.Client_secret,
			AUTH_CALLBACK_PATH+"/facebook"),
	)

	//initialize the gothic store.
	gothic.Store = sessions.NewCookieStore([]byte(conf.Session_secret))
	gothic.Store.(*sessions.CookieStore).Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * sessionDurationMinutes,
		HttpOnly: true,
		Secure:   true,
	}

	router.Get(authCallbackRelativePath+"/{provider}", authCallbackHandler)
	router.Get("/auth/{provider}", gothic.BeginAuthHandler)
	router.Delete("/logout", logoutHandler)
	router.Get("/", authHandler)
}
