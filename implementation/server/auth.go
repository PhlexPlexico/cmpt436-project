/*
 * Uses goth/gothic for authentication, and also makes use of the session that
 * gothic uses (so that there are not two sessions being used.)
 */
package server

import (
	"../db"
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
	userKey                  string = "key_user"
	authCallbackRelativePath        = "/oauth2callback"
	sessionName                     = "userSession"
)

var Store *sessions.CookieStore

type authUser struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Id        string `json:"id"` //This is the bson.ObjectId from the db.
	AvatarUrl string `json:"avatar_url"`
}

func marshalUser(user *authUser) (string, error) {
	b, err := json.Marshal(user)
	return string(b), err
}

func unmarshalUser(data string) (*authUser, error) {
	user := &authUser{}
	err := json.Unmarshal([]byte(data), user)
	return user, err
}

func getUserFromSession(s *sessions.Session) (*authUser, error) {
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
func putUserInSession(user *authUser, s *sessions.Session) error {
	userString, err := marshalUser(user)
	if err != nil {
		return err
	}
	s.Values[userKey] = userString
	return nil
}

/**
 * Validate the user for this request. If it is invalid, serve a new login.
 * @return the user, if valid, or nil if serving a new login
 */
func validateUserAndLogInIfNecessary(
	w http.ResponseWriter, r *http.Request) *authUser {
	user, err := validateUser(w, r)
	if user == nil {
		if err != nil {
			log.Println(err.Error() + ". Serving new login instead.")
		}
		serveNewLogin(w, r)
		return nil
	}

	return user
}

/**
 * return a user pointer. It is nil if the user could not be validated
 * (and thus the user is unauthorized). An error is also returned, if one
 * exists.
 */
func validateUser(
	w http.ResponseWriter, r *http.Request) (*authUser, error) {
	log.Println("validating user...")
	session, err := getSession(r)
	if err != nil {
		return nil, errors.New("unable to get user session: " + err.Error())
	}

	if session.IsNew {
		return nil, nil
	}

	user, err := getUserFromSession(session)
	if err != nil {
		return nil, errors.New("unable to unmarshal user from session: " +
			err.Error())
	}

	err = db.ValidateUser(user.Id)
	if err != nil {
		return nil, errors.New("session-stored user does not exist in database: " +
			err.Error())
	}

	return user, nil
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth request")

	if user := validateUserAndLogInIfNecessary(w, r); user != nil {
		serveIndexTemplate(w, user)
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth request")

	if user := validateUserAndLogInIfNecessary(w, r); user != nil {
		http.Redirect(w, r, "/app", http.StatusMovedPermanently)
	}
}

func serveNewLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("serving new login.")
	session, err := getSession(r)
	log.Println("ending session")
	endSession(session, w, r)

	t, err := template.New("login").Parse(loginTemplate)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	t.Execute(w, nil)
}

func serveIndexTemplate(w http.ResponseWriter, user *authUser) {
	indexTemplate := template.Must(template.ParseFiles("app/index.html"))
	err := indexTemplate.Execute(w, user)
	if err != nil {
		log.Println("error rendering template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	userId, err := db.CreateUserIfNecessary(user.Email, user.Name, user.AvatarURL, false)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		endSession(session, w, r)
		return
	}

	newAuthUser := &authUser{
		Name:      user.Name,
		Email:     user.Email,
		AvatarUrl: user.AvatarURL,
		Id:        userId,
	}

	err = putUserInSession(newAuthUser, session)
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
	endSession(session, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func endSession(s *sessions.Session, w http.ResponseWriter, r *http.Request) {
	s.Options = &sessions.Options{MaxAge: -1}
	s.Save(r, w)
}

func getSession(r *http.Request) (*sessions.Session, error) {
	return Store.Get(r, sessionName)
}

//TODO add more providers
var loginTemplate = `
<p><a href="/auth/gplus">Log in with Google</a></p>
<p><a href="/auth/facebook">Log in with Facebook</a></p>
`

type genericAuthConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func initAuth(router *pat.Router, conf *config) {
	//get all the providers set up.
	//I need "profile", "email", scopes. gplus and facebook provide these by
	//default.
	AUTH_CALLBACK_PATH := "https://" + conf.WebsiteUrl + conf.HttpsPortNum + authCallbackRelativePath
	goth.UseProviders(
		gplus.New(conf.Gplus.ClientId, conf.Gplus.ClientSecret,
			AUTH_CALLBACK_PATH+"/gplus"),
		facebook.New(conf.Facebook.ClientId, conf.Facebook.ClientSecret,
			AUTH_CALLBACK_PATH+"/facebook"),
	)

	//initialize the gothic store.
	Store = sessions.NewCookieStore([]byte(conf.Session_secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * sessionDurationMinutes,
		HttpOnly: true,
		Secure:   true,
	}

	router.Get(authCallbackRelativePath+"/{provider}", authCallbackHandler)
	router.Get("/auth/{provider}", gothic.BeginAuthHandler)
	router.Post("/logout", logoutHandler)
	router.Get("/", authHandler)
}
