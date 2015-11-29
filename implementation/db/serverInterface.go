package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

const (
	FEEDITEM_TYPE_COMMENT      string = "comment"
	FEEDITEM_TYPE_NOTIFICATION string = "notification"
	FEEDITEM_TYPE_PURCHASE     string = "purchase"
	FEEDITEM_TYPE_PAYMENT      string = "payment"
)

type FeedItem struct {
	// The actual feed item to be unmarshaled, based upon the type.
	Content *json.RawMessage `json:"content"`
	// Either a non-empty group id or contact id will be provided, but not both.
	// These are just string representations of bson.ObjectIds.
	GroupId string `json:"group_id"`
	// ContactsId string `json:"contact_id"`
	Type string `json:"type"`
}

func (fi *FeedItem) String() string {

	return fmt.Sprint(fi.GroupId, ":", fi.Type, ":", string(*fi.Content))
}

func HandleFeedItem(fi *FeedItem) error {
	switch fi.Type {
	case FEEDITEM_TYPE_COMMENT:
		comment := &Comment{}
		err := json.Unmarshal(*fi.Content, comment)
		if err != nil {
			return err
		}
		/* handler code here, verifying the contents, and
		 * doing DB insertions, etc. There is no need to rebroadcast the
		 * new feed item, as the webserver handles that automatically.
		 */
		return nil
	case FEEDITEM_TYPE_NOTIFICATION:
		return nil
	case FEEDITEM_TYPE_PAYMENT:
		return nil
	case FEEDITEM_TYPE_PURCHASE:
		return nil
	default:
		return errors.New(fmt.Sprint("invalid FeedItem type: ", fi.Type))
	}
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
