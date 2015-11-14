/*
This is just gothic, adapted by William van der Kamp to include the provider in the session store,
so that provider.FetchUser() can be called immediately, rather than going through
the whole login process.

See https://github.com/markbates/goth/blob/master/gothic/gothic.go for the source.
*/

/*
Package gothic wraps common behaviour when using Goth. This makes it quick, and easy, to get up
and running with Goth. Of course, if you want complete control over how things flow, in regards
to the authentication process, feel free and use Goth directly.
See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
)

// SESSION_NAME is the key used to access the session store.
const (
	GOTH_SESS_KEY                   = "goth_sess"
	PROVIDER_NAME_KEY               = "goth_provider"
	USER_KEY                        = "goth_user"
	SESSION_SECRET_CONFIG_FILE_PATH = "../../../.session_secret"
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

func getProviderNameFromSession(s *sessions.Session) (string, error) {
	val := s.Values[PROVIDER_NAME_KEY]
	if val == nil {
		return "", errors.New("provider not stored in session")
	}
	return val.(string), nil
}

func getProviderFromSession(s *sessions.Session) (goth.Provider, error) {
	providerName, err := getProviderNameFromSession(s)
	if err != nil {
		return nil, err
	}
	return goth.GetProvider(providerName)
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
func putProviderInSession(provider goth.Provider, s *sessions.Session) {
	s.Values[PROVIDER_NAME_KEY] = provider.Name()
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

// Store can/should be set by applications using gothic. The default is a cookie store.
var Store *sessions.CookieStore
var defaultStore *sessions.CookieStore

var keySet = false

func init() {
	//Changed by William: use a config file instead of an environment variable
	key, err := ioutil.ReadFile(SESSION_SECRET_CONFIG_FILE_PATH)
	if err != nil {
		log.Println("could not load session secret from file ",
			SESSION_SECRET_CONFIG_FILE_PATH)
		log.Println(err)

		//TODO delete these lines, when I know this will never be used.
		// key, err = os.Getenv("SESSION_SECRET")
		// if err != nil {
		// 	log.Println("could not load session secret from " +
		// 		"environment variable SESSION_SECRET.\nUsing a hard-coded value. " +
		// 		"This should be removed from production code.")
		// 	log.Println(err)
		// 	key = "aFakeTemporarySecret"
		// }
	}
	keySet = len(key) != 0
	Store = sessions.NewCookieStore([]byte(key))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * SESSION_DURATION_MINUTES,
		HttpOnly: true,
		Secure:   true,
	}
	defaultStore = Store
}

/*
BeginAuthHandler is a convienence handler for starting the authentication process.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".
BeginAuthHandler will redirect the user to the appropriate authentication end-point
for the requested provider.
See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
func BeginAuthHandler(res http.ResponseWriter, req *http.Request) {
	url, err := GetAuthURL(res, req)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, err)
		return
	}

	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

// GetState gets the state string associated with the given request
// This state is sent to the provider and can be retrieved during the
// callback.
var GetState = func(req *http.Request) string {
	return "state"
}

/*
GetAuthURL starts the authentication process with the requested provided.
It will return a URL that should be used to send users to.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".
I would recommend using the BeginAuthHandler instead of doing all of these steps
yourself, but that's entirely up to you.
*/
func GetAuthURL(res http.ResponseWriter, req *http.Request) (string, error) {

	if !keySet && defaultStore == Store {
		fmt.Println("William says: the following error should never occur!")
		fmt.Println("goth/gothic: no SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.")
	}

	providerName, err := GetProviderName(req)
	if err != nil {
		return "", err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(GetState(req))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	session, _ := Store.Get(req, SESSION_NAME)

	//added by William: store both the provider and the session.
	session.Values[GOTH_SESS_KEY] = sess.Marshal()
	session.Values[PROVIDER_NAME_KEY] = provider.Name()
	err = session.Save(req, res)
	if err != nil {
		return "", err
	}

	return url, err
}

/*
CompleteUserAuth does what it says on the tin. It completes the authentication
process and fetches all of the basic information about the user from the provider.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".
See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
var CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (goth.User, error) {

	if !keySet && defaultStore == Store {
		fmt.Println("William says: the following error should never occur!")
		fmt.Println("goth/gothic: no SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.")
	}

	providerName, err := GetProviderName(req)
	if err != nil {
		return goth.User{}, err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	session, _ := Store.Get(req, SESSION_NAME)

	if session.Values[GOTH_SESS_KEY] == nil {
		return goth.User{}, errors.New("could not find a matching session for this request")
	}

	sess, err := provider.UnmarshalSession(session.Values[GOTH_SESS_KEY].(string))
	if err != nil {
		return goth.User{}, err
	}

	_, err = sess.Authorize(provider, req.URL.Query())

	//Save the sess to the session, because now its access token has been set.
	session.Values[GOTH_SESS_KEY] = sess.Marshal()
	session.Save(req, res)

	if err != nil {
		return goth.User{}, err
	}

	return provider.FetchUser(sess)
}

// GetProviderName is a function used to get the name of a provider
// for a given request. By default, this provider is fetched from
// the URL query string. If you provide it in a different way,
// assign your own function to this variable that returns the provider
// name for your request.
var GetProviderName = getProviderName

func getProviderName(req *http.Request) (string, error) {
	log.Println("getting provider name")
	provider := req.URL.Query().Get("provider")
	if provider == "" {
		provider = req.URL.Query().Get(":provider")
	}
	if provider == "" {
		session, err := Store.Get(req, SESSION_NAME)
		if err != nil {
			return provider, err
		}

		provider, err = getProviderNameFromSession(session)
		if err != nil {
			return provider, err
		}
	}
	if provider == "" {
		return provider, errors.New("you must select a provider")
	}

	return provider, nil
}
