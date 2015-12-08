package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

const (
	FeedItemTypeComment          string = "comment"
	FeedItemTypeNotification     string = "notification"
	FeedItemTypePurchase         string = "purchase"
	FeedItemTypePayment          string = "payment"
	invalidBsonIdHexErrorMessage string = "invalid bson id hex representation"
	contactsGroupName                   = ""
)

/* For debug purposes. */
func (fi *FeedItem) String() string {
	return fmt.Sprint(fi.GroupID, ":", fi.Type, ":", string(fi.Content))
}

type FeedItemContent interface {
	TypeString() string
}

func (c *Comment) TypeString() string {
	return FeedItemTypeComment
}

func (c *Notification) TypeString() string {
	return FeedItemTypeNotification
}

func (c *Purchase) TypeString() string {
	return FeedItemTypePurchase
}

func (c *Payment) TypeString() string {
	return FeedItemTypePayment
}

/*
 * Handle all inbound websocket messages. This is where all payments and purchases
 * are given to the back-end, to be processed appropriately. There is no need for
 * this handle function to return a value to the webserver, because the webserver
 * rebroadcasts incoming feed items on its own, automatically. Just return an error
 * if the webserver should not be rebroadcasting: e.g. if an invalid purchase is made.
 *
 * This function is passed in a userId, which it uses to fill in the FeedItem, if
 * necessary. It then returns a new, filled-in feeditem.
 */
func HandleFeedItem(fi *FeedItem, userId string) (*FeedItem, error) {
	if !bson.IsObjectIdHex(fi.GroupID) || !bson.IsObjectIdHex(userId) {
		return nil, errors.New(invalidBsonIdHexErrorMessage)
	}
	switch fi.Type {
	case FeedItemTypeComment:
		comment := &Comment{}
		err := json.Unmarshal(fi.Content, comment)
		if err != nil {
			return nil, err
		}
		comment.UserID = userId
		return InsertAsFeedItem(FeedItemContent(comment), fi.GroupID)
	case FeedItemTypeNotification:
		notification := &Notification{}
		err := json.Unmarshal(fi.Content, notification)
		if err != nil {
			return nil, err
		}
		return InsertAsFeedItem(FeedItemContent(notification), fi.GroupID)
	case FeedItemTypePayment:
		payment := &Payment{}
		err := json.Unmarshal(fi.Content, payment)
		if err != nil {
			return nil, err
		}
		payment.UserId = userId
		group, err := FindGroup(bson.ObjectIdHex(fi.GroupID))
		if err != nil {
			return nil, err
		}
		err = PayMember(group, payment.PayerID, payment.PayeeID, payment.AmountInCents)
		if err != nil {
			return nil, err
		}
		return InsertAsFeedItem(FeedItemContent(payment), fi.GroupID)
	case FeedItemTypePurchase:
		purchase := &Purchase{}
		err := json.Unmarshal(fi.Content, purchase)
		if err != nil {
			return nil, err
		}
		purchase.UserId = userId
		group, err := FindGroup(bson.ObjectIdHex(fi.GroupID))
		if err != nil {
			return nil, err
		}
		err = DoPurchase(group, purchase.PayerID,
			purchase.AmountInCents, purchase.Expected)
		if err != nil {
			return nil, err
		}
		return InsertAsFeedItem(FeedItemContent(purchase), fi.GroupID)
	default:
		return nil, errors.New(fmt.Sprint("invalid FeedItem type: ", fi.Type))
	}
}

/* Return the new feed item, to be rebroadcast, with userId filled in. */
func InsertAsFeedItem(v FeedItemContent, groupId string) (*FeedItem, error) {
	if !bson.IsObjectIdHex(groupId) {
		return nil, errors.New(invalidBsonIdHexErrorMessage)
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	fi := &FeedItem{
		Content:   bytes,
		GroupID:   groupId,
		Type:      v.TypeString(),
		Timestamp: time.Now(),
	}

	// log.Printf("\n\n FeedItem %v \n \n \n", fi)
	return fi, AddFeedItemToGroupByID(bson.ObjectIdHex(groupId), fi)
}

/*
 * NOTE: the following stubs usually pass in and ask for strings instead of
 * bson.ObjectIDs. I think these can just be type-converted directly. I just
 * worked with strings in my code to... reduce coupling, I guess?
 */

/*
 * return a non-nil error if the user is not valid.
 */
func ValidateUser(userId string) error {
	if !bson.IsObjectIdHex(userId) {
		return errors.New(invalidBsonIdHexErrorMessage)
	}
	_, err := FindUserByID(bson.ObjectIdHex(userId))
	return err
}

/**
 * creates a new user if necessary, and returns a string representation of
 * the user's id. If the user already exists, just return the userId anyway.
 * If the error is not nil, the returned value must be ignored.
 */
func CreateUserIfNecessary(
	email, name, avatarUrl string, isRealUser bool) (string, error) {
	userID, err := FindUserIdByEmail(email)

	if userID == "" {
		AddUser(name, email, "", avatarUrl, isRealUser)
		userID, err = FindUserIdByEmail(email)
	}

	return userID.Hex(), err
}

/*
 * Get all groups associated with this user.
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroups(userId string) ([]Group, error) {
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New(invalidBsonIdHexErrorMessage)
	}

	user, err := FindUserByID(bson.ObjectIdHex(userId))
	if err != nil {
		return nil, err
	}

	groups := make([]Group, len(user.Groups))
	for i, groupId := range user.Groups {
		group, err := FindGroup(bson.ObjectIdHex(groupId))
		if err != nil {
			return nil, err
		}
		groups[i] = *group
	}

	return groups, nil
}

/*
 * Get the user corresponding to each userId.
 * If the error is not nil, the returned value must be ignored.
 */
func GetUsers(userIds []string) ([]User, error) {
	users := make([]User, len(userIds))
	for i, userId := range userIds {
		if !bson.IsObjectIdHex(userId) {
			return nil, errors.New(invalidBsonIdHexErrorMessage)
		}
		user, err := FindUserByID(bson.ObjectIdHex(userId))
		if err != nil {
			return nil, err
		}
		users[i] = *user
	}

	return users, nil
}

/*
 * Get all the groupIds associated with this userId.
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroupIdStrings(userId string) ([]string, error) {
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New(invalidBsonIdHexErrorMessage)
	}

	user, err := FindUserByID(bson.ObjectIdHex(userId))
	if err != nil {
		return nil, err
	}

	return user.Groups, nil
}

/*
 * Add the user with the given userId to the group with the given groupId.
 * If the error is not nil, the returned value must be ignored.
 */
func AddUsersToGroup(userIds []string, groupId string) error {
	if !bson.IsObjectIdHex(groupId) {
		return errors.New(invalidBsonIdHexErrorMessage)
	}

	for _, userId := range userIds {
		if !bson.IsObjectIdHex(userId) {
			return errors.New(invalidBsonIdHexErrorMessage)
		}
		err := AddMemberToGroupByID(bson.ObjectIdHex(groupId), bson.ObjectIdHex(userId))
		if err != nil {
			return err
		}
		err = AddGroupToUser(bson.ObjectIdHex(userId), bson.ObjectIdHex(groupId))
		if err != nil {
			return err
		}
	}

	return nil
}

/*
 * Create a group with the given group name, and with the given users as
 * members. Return the new group's ID, in string form.
 * If the error is not nil, the returned value must be ignored.
 */
func CreateGroup(name string, userIds []string) (string, error) {
	userIdsObj := make([]bson.ObjectId, len(userIds))
	for i, userIdString := range userIds {
		if !bson.IsObjectIdHex(userIdString) {
			return "", errors.New(invalidBsonIdHexErrorMessage)
		}

		userIdsObj[i] = bson.ObjectIdHex(userIdString)
	}
	log.Println()
	//Fake a group creator.
	groupId, err := AddGroup(name, userIdsObj[0])
	if err != nil {
		return "", err
	}
	log.Println()

	for _, userIdObj := range userIdsObj[1:] {
		err = AddMemberToGroupByID(groupId, userIdObj)
		if err != nil {
			return "", err
		}
		log.Println()
		err = AddGroupToUser(userIdObj, groupId)
		if err != nil {
			return "", err
		}
	}

	log.Println()
	return groupId.Hex(), nil
}

/*
 * Return the group object for the given group name (or an error if it
 * does not exist).
 * If the error is not nil, the returned value must be ignored.
 */
func GetGroup(groupId string) (*Group, error) {
	if !bson.IsObjectIdHex(groupId) {
		return nil, errors.New(invalidBsonIdHexErrorMessage)
	}

	group, err := FindGroup(bson.ObjectIdHex(groupId))
	if err != nil {
		return nil, err
	}
	return group, nil
}

/*
 * Add a contact with the given email to the given user's list of contacts.
 * Return the new Contact object.
 * If the error is not nil, the returned value must be ignored.
 */
func AddContact(userId string, contactEmail string) (string, error) {
	if !bson.IsObjectIdHex(userId) {
		return "", errors.New(invalidBsonIdHexErrorMessage)
	}

	contactId, err := FindUserIdByEmail(contactEmail)
	if err != nil {
		return "", err
	}
	log.Println()
	groupIdString, err := CreateGroup(
		contactsGroupName, []string{userId, contactId.Hex()})
	if err != nil {
		return "", err
	}
	log.Println()
	return groupIdString, nil
}

/*
 * Get the list of all feed items for the given group,
 * ordered in ascending order of timestamp.
 * If the error is not nil, the returned value must be ignored.
 */
func GetAllFeedItems(groupId bson.ObjectId) ([]FeedItem, error) {
	group, err := FindGroup(groupId)
	if err != nil {
		return nil, err
	}

	return group.Feed, nil
}
