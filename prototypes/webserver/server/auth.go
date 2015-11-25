/*
 * Uses goth/gothic for authentication, and also makes use of the session that
 * gothic uses (so that there are not two sessions being used.)
 */
package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"io/ioutil"
	"log"
)

// SESSION_NAME is the key used to access the session store.
const (
	USER_KEY                        string = "goth_user"
	SESSION_DURATION_MINUTES        int    = 30
	SESSION_SECRET_CONFIG_FILE_PATH string = "../../../.session_secret"
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
	val := s.Values[USER_KEY]
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
	s.Values[USER_KEY] = userString
	return nil
}

func initStore() {
	key, err := ioutil.ReadFile(SESSION_SECRET_CONFIG_FILE_PATH)
	if err != nil {
		log.Println("could not load session secret from file ",
			SESSION_SECRET_CONFIG_FILE_PATH)
		log.Fatalln(err)
	}
	gothic.Store = sessions.NewCookieStore([]byte(key))
	gothic.Store.(*sessions.CookieStore).Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * SESSION_DURATION_MINUTES,
		HttpOnly: true,
		Secure:   true,
	}
}
