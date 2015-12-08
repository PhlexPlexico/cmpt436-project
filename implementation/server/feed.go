package server

import (
	"../db"
	"encoding/json"
	"fmt"
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
	Balance   int    `json:"balance"`
}

type uiGroup struct {
	Name      string        `json:"name"`
	Id        string        `json:"id"`
	Users     []uiUser      `json:"users"`
	FeedItems []db.FeedItem `json:"feed_items"`
}

func NewFeedsManager() *feedsManager {
	fm := &feedsManager{
		join:            make(chan *connection),
		leave:           make(chan *connection),
		incoming:        make(chan *db.FeedItem),
		addToGroup:      make(chan *userIdsGroupId),
		clients:         make(map[string]*connection),
		clientsPerGroup: make(map[string]map[string]*connection),
	}
	fm.listen()
	return fm
}

func (fm *feedsManager) listen() {
	go func() {
		for {
			fm.printState()
			select {
			case uidsAndGid := <-fm.addToGroup:
				fm.addNewClientsToFeedById(uidsAndGid.userIds, uidsAndGid.groupId,
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

	//Get all groups to which this user belong.
	groups, err := db.GetGroups(client.userId)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//Give the client all group data up to this point.
	uiGroups := make([]uiGroup, len(groups))
	for i, group := range groups {
		newUiGroup, err := createUiGroup(&group)
		if err != nil {
			log.Println(err)
			return
		}
		uiGroups[i] = *newUiGroup
	}
	uiGroupsBytes, err := json.Marshal(uiGroups)
	if err != nil {
		log.Println(err.Error())
		return
	}
	client.outgoing <- &websocketOutMessage{
		Content: uiGroupsBytes,
		Type:    messageTypeGroups,
	}

	//register the client for notifications from each of its groups.
	for _, group := range groups {
		fm.addClientToFeed(client, group.ID.Hex(), fm.clientsPerGroup)
	}

	fmt.Printf("client joined. client's groups sent:\n%v\n\n", uiGroups)
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
func (fm *feedsManager) addNewClientsToFeedById(userIds []string, feedId string,
	feeds map[string]map[string]*connection) {
	group, err := db.GetGroup(feedId)
	if err != nil {
		log.Println(err)
		return
	}

	//Send the new users to all the older users in the group.
	uiUsers := make([]uiUser, len(userIds))
	for i, userId := range userIds {
		if !bson.IsObjectIdHex(userId) {
			log.Println("invalid format for feedId.")
			return
		}
		user, err := db.FindUserByID(bson.ObjectIdHex(userId))
		if err != nil {
			log.Println(err.Error())
			return
		}
		uiUsers[i] = *createUiUser(user, group.Actual[i])
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

	notifs := make([]*db.FeedItem, len(userIds))
	var wsMessage *websocketOutMessage
	for i, userId := range userIds {
		if client, ok := fm.clients[userId]; ok {
			//If the added client is active, send them their new group.
			if wsMessage == nil {

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

		//Create a notification associated with the new user.
		notification := &db.Notification{
			Content: uiUsers[i].Name + "joined the group.",
		}
		err = db.InsertAsFeedItem(db.FeedItemContent(notification), feedId)
		if err != nil {
			log.Println(err.Error())
			return
		}
		notificationBytes, err := json.Marshal(notification)
		if err != nil {
			log.Println(err.Error())
			return
		}
		notifs[i] = &db.FeedItem{
			Content: notificationBytes,
			GroupID: feedId,
			Type:    db.FeedItemTypeNotification,
		}
	}

	//Notify the feed (including the new users) of the new users in the group.
	for _, notif := range notifs {
		fm.broadcastFeedItem(notif)
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
		uiUsers[j] = *createUiUser(&user, group.Actual[j])
	}
	feedItems, err := db.GetAllFeedItems(group.ID)
	if err != nil {
		return nil, err
	}
	return &uiGroup{
		Name:      group.GroupName,
		Id:        group.ID.Hex(),
		Users:     uiUsers,
		FeedItems: feedItems,
	}, nil
}

func createUiUser(user *db.User, balance int) *uiUser {
	return &uiUser{
		Name:      user.Name,
		Id:        user.ID.Hex(),
		AvatarUrl: user.AvatarURL,
		Balance:   balance,
	}
}

func (fm *feedsManager) printState() {
	fmt.Printf("\n####### Current Feed Manager State #######\n"+
		"active clients: %v\n\nactive groups: %v\n\n"+
		"##########################################\n\n",
		fm.clients, fm.clientsPerGroup)
}
