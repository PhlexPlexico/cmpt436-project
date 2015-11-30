package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

const (
	feedItemTypeComment      string = "comment"
	feedItemTypeNotification string = "notification"
	feedItemTypePurchase     string = "purchase"
	feedItemTypePayment      string = "payment"
)

/*
 * All inbound websocket messages take this form.
 */
type FeedItem struct {
	// The actual feed item to be unmarshaled, based upon the type.
	Content json.RawMessage `json:"content"`
	// This is just a string representation of a bson.ObjectId.
	GroupId string `json:"group_id"`
	// ContactsId string `json:"contact_id"`
	Type string `json:"type"`
}

/* For debug purposes. */
func (fi *FeedItem) String() string {
	return fmt.Sprint(fi.GroupId, ":", fi.Type, ":", string(fi.Content))
}

/*
 * Handle all inbound websocket messages. This is where all payments and purchases
 * are given to the back-end, to be processed appropriately. There is no need for
 * this handle function to return a value to the webserver, because the webserver
 * rebroadcasts incoming feed items on its own, automatically. Just return an error
 * if the webserver should not be rebroadcasting: e.g. if an invalid purchase is made.
 */
func HandleFeedItem(fi *FeedItem) error {
	switch fi.Type {
	case feedItemTypeComment:
		comment := &Comment{}
		err := json.Unmarshal(fi.Content, comment)
		if err != nil {
			return err
		}
		err = comment.Insert()
		if err != nil {
			return err
		}

	case feedItemTypeNotification:
		notification := &Notification{}
		err := json.Unmarshal(fi.Content, notification)
		if err != nil {
			return err
		}
		err = notification.Insert()
		if err != nil {
			return err
		}
	case feedItemTypePayment:
		payment := &Payment{}
		err := json.Unmarshal(fi.Content, payment)
		if err != nil {
			return err
		}
		err = payment.Insert()
		if err != nil {
			return err
		}
	case feedItemTypePurchase:
		purchase := &Purchase{}
		err := json.Unmarshal(fi.Content, purchase)
		if err != nil {
			return err
		}
		err = purchase.Insert()
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprint("invalid FeedItem type: ", fi.Type))
	}

	return nil
}

func (c *Comment) Insert() error {
	return nil
}

func (n *Notification) Insert() error {
	return nil
}

func (pu *Purchase) Insert() error {
	return nil
}

func (pa *Payment) Insert() error {
	return nil
}

/*
 * NOTE: the following stubs usually pass in and ask for strings instead of
 * bson.ObjectIDs. I think these can just be type-converted directly. I just
 * worked with strings in my code to... reduce coupling, I guess?
 */

/**
 * creates a new user if necessary, and returns a string representation of
 * the user's id. If the user already exists, just return the userId anyway.
 * If the error is not nil, the returned value must be ignored.
 */
func CreateUserIfNecessary(
	email, name, avatarUrl string, isRealUser bool) (string, error) {
	// func GetUserIdString(email string) (string, error) {
	/* userID := GetIDbyEmail(email)
	 * if userID == "" {
	 * 	AddUser("", email, "", 0)
	 *  userID = GetIDbyEmail(email)
	 *  return userID
	 * }
	 * return userID, nil
	 */
	return "", nil
}

/*
 * Get all groups associated with this user.
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroups(userId string) ([]Group, error) {
	return nil, nil
}

/*
 * Get the user corresponding to each userId.
 * If the error is not nil, the returned value must be ignored.
 */
func GetUsers(userIds []bson.ObjectId) ([]User, error) {
	return nil, nil
}

/*
 * Get all the groupIds associated with this userId.
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroupIdStrings(userId string) ([]string, error) {
	return nil, nil
}

/*
 * Add the user with the given userId to the group with the given groupId.
 * If the error is not nil, the returned value must be ignored.
 */
func AddUserToGroup(userId, groupId string) error {
	return nil
}

/*
 * Create a group with the given group name, and with the given users as
 * members. Return the new group's ID, in string form.
 * If the error is not nil, the returned value must be ignored.
 */
func CreateGroup(name string, userIds []string) (string, error) {
	return "", nil
}

/*
 * Return the group object for the given group name (or an error if it
 * does not exist).
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroup(groupId string) (*Group, error) {
	return nil, nil
}

/*
 * Add a contact with the given email to the given user's list of contacts.
 * Return the new Contact object.
 * If the error is not nil, the returned value must be ignored.
 */
func AddContact(userId string, contactEmail string) (*Contact, error) {
	return nil, nil
}

/*
 * Get the list of all feed items for the given group,
 * ordered in ascending order of timestamp.
 * If the error is not nil, the returned value must be ignored.
 */
func GetAllFeedItems(groupId bson.ObjectId) ([]FeedItem, error) {
	return nil, nil
}
