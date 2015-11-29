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

type FeedItem struct {
	// The actual feed item to be unmarshaled, based upon the type.
	Content json.RawMessage `json:"content"`
	// Either a non-empty group id or contact id will be provided, but not both.
	// These are just string representations of bson.ObjectIds.
	GroupId string `json:"group_id"`
	// ContactsId string `json:"contact_id"`
	Type string `json:"type"`
}

func (fi *FeedItem) String() string {

	return fmt.Sprint(fi.GroupId, ":", fi.Type, ":", string(fi.Content))
}

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
	case feedItemTypeNotification:
	case feedItemTypePayment:
	case feedItemTypePurchase:
	default:
		return errors.New(fmt.Sprint("invalid FeedItem type: ", fi.Type))
	}

	return nil
}

/**
 * creates a new user if necessary, and returns a string representation of
 * the user's id.
 */
func GetUserIdString(email string) (string, error) {
	return "", nil
}

func GetGroups(userId string) ([]Group, error) {
	return nil, nil
}

/* Get the user corresponding to each userId. */
func GetUsers(userIds []bson.ObjectId) ([]User, error) {
	return nil, nil
}

func GetGroupIdStrings(userId string) ([]string, error) {
	return nil, nil
}

func AddUserToGroup(userId, groupId string) error {
	return nil
}

func CreateGroup(name string, userIds []string) (string, error) {
	return "", nil
}

func AddContact(userId string, contactEmail string) (Contact, error) {
	return Contact{}, nil
}
