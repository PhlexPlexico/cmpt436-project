package db

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

const (
	FeedItemTypeComment      string = "comment"
	FeedItemTypeNotification string = "notification"
	FeedItemTypePurchase     string = "purchase"
	FeedItemTypePayment      string = "payment"
)

/* For debug purposes. */
func (fi *FeedItem) String() string {
	return fmt.Sprint(fi.GroupID, ":", fi.Type, ":", string(fi.Content))
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
	case FeedItemTypeComment:
		comment := &Comment{}
		err := json.Unmarshal(fi.Content, comment)
		if err != nil {
			return err
		}
		err = comment.Insert()
		if err != nil {
			return err
		}

	case FeedItemTypeNotification:
		notification := &Notification{}
		err := json.Unmarshal(fi.Content, notification)
		if err != nil {
			return err
		}
		err = notification.Insert()
		if err != nil {
			return err
		}
	case FeedItemTypePayment:
		payment := &Payment{}
		err := json.Unmarshal(fi.Content, payment)
		if err != nil {
			return err
		}
		err = payment.Insert()
		if err != nil {
			return err
		}
	case FeedItemTypePurchase:
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
	/* userID := FindUserIdByEmail(email)
	 * if userID == "" {
	 * 	AddUser(name, email, "", avatarUrl, isRealUser)
	 *  userID = FindUserIdByEmail(email)
	 *  return userID
	 * }
	 * return userID, nil
	 */
	return "", nil
}

func CreateNotification(subject, content, groupId string) (Notification, error) {
	return Notification{}, nil
}

/*
 * Get all groups associated with this user.
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroups(userId string) ([]Group, error) {
	/*
		var err error
		var user User
		//newUID := bson.ObjectId(userId)
		user, err = FindUserByID(bson.ObjectId(userId))
		if err != nil {
			return err
		}
		return user
		**/
	return nil, nil
}

/*
 * Get the user corresponding to each userId.
 * If the error is not nil, the returned value must be ignored.
 */
func GetUsers(userIds []string) ([]User, error) {
	/*
		var users[] User
		var err error
		for _, i := range userIds {
			users[i], err = FindUserByID(bson.ObjectId(userIds[i]))
			if err != nil {
				return err
			}
		}
		return users, nil
	*/
	return nil, nil
}

/*
 * Get all the groupIds associated with this userId.
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroupIdStrings(userId string) ([]string, error) {
	/*
		var user User
		var err error
		user, err = FindUserByID(bson.ObjectId(userId))
		if err != nil {
			return nil, err
		}
		return user.Groups, nil
	*/
	return nil, nil
}

/*
 * Add the user with the given userId to the group with the given groupId.
 * If the error is not nil, the returned value must be ignored.
 */
func AddUsersToGroup(userIds []string, groupId string, adderId string) error {
	/*
		var err error
		for _, i := range userIds {
			err = AddMemberToGroupByID(bson.ObjectId(groupId), bson.ObjectId(userIds[i]))
			if err != nil {
				return err
			}
			err = AddGroupToUser(bson.ObjectId(userIds[i]), bson.ObjectId(groupId))
			if err != nil {
				return err
			}
		}
		return nil

	*/
	return nil
}

/*
 * Create a group with the given group name, and with the given users as
 * members. Return the new group's ID, in string form.
 * If the error is not nil, the returned value must be ignored.
 */
func CreateGroup(name string, userIds []string) (string, error) {
	/*
			var err error
			var group Group
			// AddGroup adds all users to the group and in the User field as well??
			err = AddGroup(bson.ObjectId(name), bson.ObjectId(userIds))
			if err != nil {
				return "", err
			}
			user := FindUserByID(bson.ObjectId(userIds[1]))
			for _, i := range user.Groups {
				group, err = FindGroup(bson.ObjectId(user.Group[i]))
				if err != nil {
					return "", err
				}
				ifgroup.GroupName == name {
					break
				}
			}
			return group.ID.hex(), nil
		}
	*/
	return "", nil
}

/*
 * Return the group object for the given group name (or an error if it
 * does not exist).
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroup(groupId string) (*Group, error) {
	/*
		var err error
		var group Group
		group, err = FindGroup(bson.ObjectId(groupId))
		if err != nil {
			return nil, err
		}
		return group, nil
	*/
	return nil, nil
}

/*
 * Add a contact with the given email to the given user's list of contacts.
 * Return the new Contact object.
 * If the error is not nil, the returned value must be ignored.
 */
func AddContact(userId string, contactEmail string) (*Contact, error) {
	/*
		var err error
		var userCon User
		var user User
		var newContact Contact
		user, err = FindUserById(bson.ObjectId(userId))
		if err != nil {
			return nil, err
		}
		userCon, err = FindUserIdByEmail(bson.ObjectId(contactEmail))
		if err != nil {
			return nil, err
		}
		err = AddContact_other(userCon.Name, userCon.Email, userCon.Phone, userCon.isRealUser, user.ID)
		if err != nil {
			return nil, err
		}
		newContact, err = FindContact(userCon.id)
		if err != nil {
			return nil, err
		}
		return newContact, nil
	*/
	return nil, nil
}

/*
 * Get the list of all feed items for the given group,
 * ordered in ascending order of timestamp.
 * If the error is not nil, the returned value must be ignored.
 */
func GetAllFeedItems(groupId bson.ObjectId) ([]FeedItem, error) {
	/*
		var err error
		var feed []FeedItem
		feed, err = FindFeedItemByGroupId(groupId)
		if err != nil {
			return err
		}
		return feed, nil
	*/
	return nil, nil
}
