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

	body := addGroupBody{}
	err := r.DecodeJsonPayload(&body)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	if err := db.AddUsersToGroup(
		body.ContactUserIds, body.GroupId, user.Id); err == nil {
		w.WriteHeader(http.StatusOK)
		fm.addToGroup <- &userIdsGroupId{
			userIds: []string{body.ContactUserIds},
			groupId: body.groupId,
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

	if groupId, err := db.CreateGroup(body.Name, body.MemberIds); groupId != "" {
		w.WriteHeader(http.StatusOK)
		fm.addToGroup <- &userIdsGroupId{
			userIds: body.MemberIds,
			groupId: groupId,
		}
	} else if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		rest.Error(w, "unable to create group", http.StatusInternalServerError)
	}
}

// func getContactsHandler(w rest.ResponseWriter, r *rest.Request) {
// 	user := castedValidateUserAndLogInIfNecessary(w, r)
// 	if user == nil {
// 		return
// 	}

// 	if contacts, err := db.GetContacts(user.Id); contacts != nil {
// 		contactsBody := make([]uiUser, len(contacts))
// 		for i, contact := range contacts {
// 			contactsBody[i] = uiUser{
// 				Name:      contact.Name,
// 				Id:        contacts.Id,
// 				AvatarUrl: contacts.AvatarUrl,
// 			}
// 		}
// 		w.writeJson(&contactsBody)
// 	} else if err != nil {
// 		rest.Error(w, err.Error(), http.StatusInternalServerError)
// 	} else {
// 		rest.Error(w, "unable to get contacts", http.StatusInternalServerError)
// 	}
// }

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

	if contact, err := db.AddContact(user.Id, body.Email); contact.Name != "" {
		if groupId, err := db.CreateGroup("",
			[]string{user.Id, string(contact.ID)}); groupId != "" {
			w.WriteHeader(http.StatusOK)
			fm.addToGroup <- &userIdsGroupId{
				userIds: []string{user.Id},
				groupId: groupId,
			}
		} else if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			rest.Error(w, "unable to create group for contacts.",
				http.StatusInternalServerError)
		}
	} else if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		rest.Error(w, "unable to add contact", http.StatusInternalServerError)
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
	log.Fatal(http.ListenAndServe(conf.RestPortNum, api.MakeHandler()))
}
