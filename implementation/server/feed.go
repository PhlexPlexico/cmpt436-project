package server

import (
	"../db"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"log"
)

var fm *feedsManager

type feedsManager struct {
	join       chan *connection
	leave      chan *connection
	incoming   chan *db.FeedItem
	addToGroup chan *userIdsGroupId
	// addToContacts      chan *userIdsGroupId
	clients         map[string]*connection
	clientsPerGroup map[string]map[string]*connection
	// clientsPerContacts map[string]map[string]*connection
}

//This can be used for both groups and contacts, because groupId and
//contactId are the same type.
type userIdsGroupId struct {
	userIds []string
	groupId string
}

type uiUser struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	AvatarUrl string `json:"avatar_url"`
}

type uiGroup struct {
	Name      string        `json:"name"`
	Id        string        `json:"id"`
	Balances  []int         `json:"balances"`
	Users     []uiUser      `json:"users"`
	FeedItems []db.FeedItem `json:"feed_items"`
}

func NewFeedsManager() *feedsManager {
	fm := &feedsManager{
		join:       make(chan *connection),
		leave:      make(chan *connection),
		incoming:   make(chan *db.FeedItem),
		addToGroup: make(chan *userIdsGroupId),
		// addToContacts:      make(chan *userIdsGroupId),
		clients:         make(map[string]*connection),
		clientsPerGroup: make(map[string]map[string]*connection),
		// clientsPerContacts: make(map[string]map[string]*connection),
	}
	fm.listen()
	return fm
}

func (fm *feedsManager) listen() {
	go func() {
		for {
			select {
			// case uidAndGid := <-fm.addToContacts:
			// 	fm.addClientToFeedById(uidAndGid.userId, uidAndGid.groupId,
			// 		fm.clientsPerContacts)
			case uidsAndGid := <-fm.addToGroup:
				fm.addClientsToFeedById(uidsAndGid.userIds, uidsAndGid.groupId,
					fm.clientsPerGroup)
			case client := <-fm.join:
				fm.joinHandler(client)
			case client := <-fm.leave:
				fm.leaveHandler(client)
			case message := <-fm.incoming:
				if err := db.HandleFeedItem(message); err == nil {
					fm.broadcastFeedItem(message)
				} else {
					log.Println("could not handle message", message,
						",\ndue to error:", err.Error())
				}
			}
		}
	}()
}

func (fm *feedsManager) joinHandler(client *connection) {
	fm.clients[client.userId] = client

	groups, err := db.GetGroups(client.userId)
	if err != nil {
		log.Println(err.Error())
		return
	}
	uiGroups := make([]uiGroup, len(groups))
	for i, group := range groups {
		//register the client for notifications from each of its groups.
		newUiGroup, err := createUiGroup(&group)
		if err != nil {
			log.Println(err)
			return
		}
		uiGroups[i] = *newUiGroup
	}
	//Give the client all group data up to this point.
	uiGroupsBytes, err := json.Marshal(uiGroups)
	if err != nil {
		log.Println(err.Error())
		return
	}
	client.outgoing <- &websocketOutMessage{
		Content: uiGroupsBytes,
		Type:    messageTypeGroups,
	}

	//give the client all contacts data.
	//This is probably not needed.
	// contacts := db.GetContacts(client.userId)
	// uiContacts := make([]uiUser, len(contacts))
	// for i, user := range users {
	// 	uiContacts[i] = uiUser{
	// 		Name:      user.Name,
	// 		Id:        user.Id,
	// 		AvatarUrl: user.AvatarUrl,
	// 	}
	// }
	// client.outgoing <- &websocketOutMessage{
	// 	Content: uiContacts,
	// 	Type:    messageTypeContacts,
	// }

	log.Println("client joined")
}

func (fm *feedsManager) leaveHandler(client *connection) {
	if _, ok := fm.clients[client.userId]; !ok {
		log.Println("unregistered client tried to leave.")
		return
	}

	delete(fm.clients, client.userId)
	close(client.outgoing)

	groupIds, err := db.GetGroupIdStrings(client.userId)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, groupId := range groupIds {
		removeClientFromFeed(client, groupId, fm.clientsPerGroup)
	}
	// contactsIds := db.GetContactsIds(client.userId)
	// for contactsId := range contactsIds {
	// 	removeClientFromFeed(client, contactsId, fm.clientsPerContacts)
	// }

	log.Println("client left")
}

func (fm *feedsManager) broadcastFeedItem(message *db.FeedItem) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err.Error())
		return
	}
	wsMessage := &websocketOutMessage{
		Content: messageBytes,
		GroupId: message.GroupID,
		Type:    messageTypeFeedItem,
	}

	fm.broadcast(wsMessage)
	// if message.Gid != "" {

	// } else {
	// 	for client := range fm.clientsPerContacts[message.ContactsId] {
	// 		client.outgoing <- wsMessage
	// 	}
	// }
	log.Println("broadcasted message to group " + message.GroupID)
}

func (fm *feedsManager) broadcast(message *websocketOutMessage) {
	for _, client := range fm.clientsPerGroup[message.GroupId] {
		log.Println()
		client.outgoing <- message
	}
}

/*
 * This will only add the client with id userId to the broadcast for the feed
 * with feedId if the client is currently connected.
 */
func (fm *feedsManager) addClientsToFeedById(userIds []string, feedId string,
	feeds map[string]map[string]*connection) {

	uiUsers := make([]uiUser, len(userIds))
	for i, userId := range userIds {
		user, err := db.FindUserByID(bson.ObjectId(userId))
		if err != nil {
			log.Println(err.Error())
			return
		}
		uiUsers[i] = *createUiUser(&user)
	}

	uiUsersBytes, err := json.Marshal(uiUsers)
	if err != nil {
		log.Println(err.Error())
		return
	}
	fm.broadcast(&websocketOutMessage{
		Content: uiUsersBytes,
		GroupId: feedId,
		Type:    messageTypeUsers,
	})

	notifs := make([]db.Notification, len(userIds))
	var wsMessage *websocketOutMessage
	for i, userId := range userIds {
		if client, ok := fm.clients[userId]; ok {
			if wsMessage == nil {
				group, err := db.GetGroup(feedId)
				if err != nil {
					log.Println(err)
					return
				}
				newUiGroup, err := createUiGroup(group)
				if err != nil {
					log.Println(err)
					return
				}

				uiGroupsBytes, err := json.Marshal([]*uiGroup{newUiGroup})
				if err != nil {
					log.Println(err)
					return
				}
				wsMessage = &websocketOutMessage{
					Content: uiGroupsBytes,
					Type:    messageTypeGroups,
				}
			}
			fm.addClientToFeed(client, feedId, feeds)
			client.outgoing <- wsMessage
			log.Println("added client to group broadcast")
		} else {
			log.Println("client not connected; no need to add it to a new broadcast.")
		}

		var err error
		notifs[i], err = db.CreateNotification(userId, "joined the group.", feedId)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}

	for _, notif := range notifs {
		notifBytes, err := json.Marshal(notif)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fm.broadcastFeedItem(&db.FeedItem{
			Content: notifBytes,
			GroupID: feedId,
			Type:    db.FeedItemTypeNotification,
		})
	}
}

func (fm *feedsManager) addClientToFeed(client *connection, feedId string,
	feeds map[string]map[string]*connection) {
	if clientsThisFeed, exists := feeds[feedId]; exists {
		clientsThisFeed[client.userId] = client
	} else {
		feeds[feedId] = make(map[string]*connection)
		feeds[feedId][client.userId] = client
	}
}

func removeClientFromFeed(client *connection, feedId string,
	feeds map[string]map[string]*connection) {
	if clientsThisFeed, exists := feeds[feedId]; exists {
		if _, ok := clientsThisFeed[client.userId]; ok {
			delete(clientsThisFeed, client.userId)

			if len(clientsThisFeed) == 0 {
				delete(feeds, feedId)
				log.Println("unregistering feed with Id ", feedId)
			}
		}
	}
}

func createUiGroup(group *db.Group) (*uiGroup, error) {
	users, err := db.GetUsers(group.UserIDs)
	if err != nil {
		return nil, err
	}
	uiUsers := make([]uiUser, len(users))
	for j, user := range users {
		uiUsers[j] = *createUiUser(&user)
	}
	feedItems, err := db.GetAllFeedItems(group.ID)
	if err != nil {
		return nil, err
	}
	return &uiGroup{
		Name:      group.GroupName,
		Id:        string(group.ID),
		Balances:  group.Actual,
		Users:     uiUsers,
		FeedItems: feedItems,
	}, nil
}

func createUiUser(user *db.User) *uiUser {
	return &uiUser{
		Name:      user.Name,
		Id:        string(user.ID),
		AvatarUrl: user.AvatarUrl,
	}
}
