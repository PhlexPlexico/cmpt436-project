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
	// Either a non-empty group id or contact id will be provided, but not both.
	// These are just string representations of bson.ObjectIds.
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
		/* handler code here, verifying the contents, and
		 * doing DB insertions, etc. There is no need to rebroadcast the
		 * new feed item, as the webserver handles that automatically.
		 */

		/* The rest of the cases can be done exactly like the comment case. */
	case feedItemTypeNotification:
	case feedItemTypePayment:
	case feedItemTypePurchase:
	default:
		return errors.New(fmt.Sprint("invalid FeedItem type: ", fi.Type))
	}

	return nil
}

/*
 * NOTE: the following stubs usually pass in and ask for strings instead of
 * bson.ObjectIDs. I think these can just be type-converted directly. I just
 * worked with strings in my code to... reduce coupling, I guess?
 */

/**
 * creates a new user if necessary, and returns a string representation of
 * the user's id.
 * If the error is not nil, the returned value must be ignored.
 */
func GetUserIdString(email string) (string, error) {
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
 * Add a contact with the given email to the given user's list of contacts.
 * Return the new Contact object.
 * If the error is not nil, the returned value must be ignored.
 */
func AddContact(userId string, contactEmail string) (Contact, error) {
	return Contact{}, nil
}