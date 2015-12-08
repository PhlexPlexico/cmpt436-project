package server

import (
	"../db"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

type addContactBody struct {
	Email string `json:"email"`
}

type addContactsToGroupBody struct {
	ContactUserIds []string `json:"contact_user_id"`
	GroupId        string   `json:"group_id"`
}

type createGroupBody struct {
	Name      string   `json:"name"`
	MemberIds []string `json:"member_ids"`
}

func castedValidateUserAndLogInIfNecessary(
	w rest.ResponseWriter, r *rest.Request) *authUser {
	return validateUserAndLogInIfNecessary(
		w.(http.ResponseWriter), r.Request)
}
func addContactsToGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	user := castedValidateUserAndLogInIfNecessary(w, r)
	if user == nil {
		return
	}

	body := addContactsToGroupBody{}
	err := r.DecodeJsonPayload(&body)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	if err := db.AddUsersToGroup(
		body.ContactUserIds, body.GroupId); err == nil {
		w.WriteHeader(http.StatusOK)
		fm.addToGroup <- &userIdsGroupId{
			userIds: body.ContactUserIds,
			groupId: body.GroupId,
		}
	} else {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	user := castedValidateUserAndLogInIfNecessary(w, r)
	if user == nil {
		return
	}

	body := &createGroupBody{}
	err := r.DecodeJsonPayload(body)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	if groupId, err := db.CreateGroup(body.Name, body.MemberIds); err == nil {
		w.WriteHeader(http.StatusOK)
		fm.addToGroup <- &userIdsGroupId{
			userIds: body.MemberIds,
			groupId: groupId,
		}
	} else {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addContactHandler(w rest.ResponseWriter, r *rest.Request) {
	user := castedValidateUserAndLogInIfNecessary(w, r)
	if user == nil {
		return
	}

	body := addContactBody{}
	err := r.DecodeJsonPayload(&body)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	groupId, err := db.AddContact(user.Id, body.Email)
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		group, err := db.GetGroup(groupId)
		if err != nil {
			log.Println(err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			fm.addToGroup <- &userIdsGroupId{
				userIds: group.UserIDs,
				groupId: groupId,
			}
		}
	}
}

func serveRestApi(conf *config) {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		// rest.Get("/groups", getGroupsHandler),
		// rest.Get("/contacts", getContactsHandler),
		rest.Post("/addcontactstogroup", addContactsToGroupHandler),
		rest.Post("/creategroup", createGroupHandler),
		rest.Post("/addcontact", addContactHandler),
		// rest.Delete("/removegroup"),
		// rest.Delete("/removecontact"),
		// rest.Delete("/removeuser"),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))
}
