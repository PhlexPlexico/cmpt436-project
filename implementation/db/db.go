package db

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

////////////////////////////////////////////////////////
//          DATABASE SCHEMA           //
////////////////////////////////////////////////////////
type User struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	IsRealUser bool          `json:"isRealUser"`
	Groups     []string      `json:"groups"`
	Contacts   []string      `json:"contacts"`
	Timestamp  time.Time     `json:"time"`
}

type Group struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	GroupName string        `json:"groupName"`
	UserIDs   []string      `json:"userids"`
	Expected  []int         `json:"expected"`
	Actual    []int         `json:"actual"`
	Feed      []FeedItem    `json:"feedItem"`
}

type Contact struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	IsRealUser bool          `json:"isRealUser`
	Timestamp  time.Time     `json:"time"`
}

type Comment struct {
	ID        bson.ObjectId `json:"id" bson:"_id, omitempty"`
	UserName  string        `json:"userName"`
	UserID    string        `json:"userid"`
	Subject   string        `json:"subject"`
	Content   string        `json:"content"`
	Timestamp time.Time     `json:"time"`
}

type Notification struct {
	ID        bson.ObjectId `json:"id" bson:"_id, omitempty"`
	UserID    string        `json:"userid"`
	Subject   string        `json:"subject"`
	Content   string        `json:"content"`
	Timestamp time.Time     `json:"time"`
}

type Payment struct {
	ID            bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Payer         string        `json:"payer"`
	PayerID       string        `json:"payerid"`
	Payee         string        `json:"payee"`
	PayeeID       string        `json:"payeeid"`
	AmountInCents int           `json:"amountInCents"`
	Timestamp     time.Time     `json:"time"`
}

type Purchase struct {
	ID            bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Payer         string        `json:"payer"`
	UserIDs       []string      `json:"userids"`
	Expected      []int         `json:"expected"`
	AmountInCents int           `json:"amountInCents"`
	Timestamp     time.Time     `json:"time"`
}

type FeedItem struct {
	ID        bson.ObjectId   `json:"id" bson:"_id, omitempty"`
	Content   json.RawMessage `json:"content"`
	GroupID   string          `json:"groupid"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"time"`
}

var (
	IsDrop  = true
	Session *mgo.Session
	Col     *mgo.Collection
)

////////////////////////////////////////////////////////
//          USER FUNCTIONS            //
////////////////////////////////////////////////////////
func AddUser(name string, email string, phone string, isRealUser bool) error {
	var err error
	Col = Session.DB("test").C("User")
	err = Col.Insert(&User{Name: name, Phone: phone, IsRealUser: isRealUser, Email: email, Timestamp: time.Now()})
	ThisPanic(err)
	return err
}

func FindUserByID(id bson.ObjectId) (User, error) {
	var err error
	Col = Session.DB("test").C("User")
	user := User{}
	err = Col.Find(bson.M{"_id": bson.ObjectId(id)}).One(&user)
	return user, err
}

func FindUserIdByEmail(email string) (bson.ObjectId, error) {
	var err error
	Col = Session.DB("test").C("User")
	user := User{}
	err = Col.Find(bson.M{"email": email}).One(&user)
	return user.ID, err
}

func AddGroupToUser(userId bson.ObjectId, groupId bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("User")
	query := bson.M{"_id": bson.ObjectId(userId)}
	change := bson.M{"$push": bson.M{"groups": groupId.Hex()}}
	err = Col.Update(query, change)
	return err
}

func AddContactToUser(userId bson.ObjectId, contactId bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("User")
	query := bson.M{"_id": bson.ObjectId(userId)}
	change := bson.M{"$push": bson.M{"contacts": contactId.Hex()}}
	err = Col.Update(query, change)
	return err
}

////////////////////////////////////////////////////////
//          GROUP FUNCTIONS           //
////////////////////////////////////////////////////////
func AddGroup(groupName string, uid bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("Group")
	id := bson.NewObjectId()
	err = Col.Insert(&Group{ID: id, GroupName: groupName, UserIDs: []string{uid.Hex()}, Expected: []int{0}, Actual: []int{0}})
	AddGroupToUser(uid, id)
	return err
}

func FindGroup(id bson.ObjectId) (Group, error) {
	var err error
	Col = Session.DB("test").C("Group")
	group := Group{}
	err = Col.Find(bson.M{"_id": bson.ObjectId(id)}).One(&group)
	ThisPanic(err)
	return group, err
}

func AddMemberToGroupByID(groupId bson.ObjectId, userId bson.ObjectId) error {
	var err error
	g, err := FindGroup(groupId)
	Col = Session.DB("test").C("Group")
	query := bson.M{"_id": g.ID}
	change := bson.M{"$push": bson.M{"userids": userId.Hex(), "expected": 0, "actual": 0}}
	err = Col.Update(query, change)
	return err
}

func GetGroupChanges(g Group) error {
	var err error
	Col = Session.DB("test").C("Group")
	query := bson.M{"_id": g.ID}
	change := bson.M{"$set": bson.M{"groupName": g.GroupName, "users": g.UserIDs, "expected": g.Expected, "actual": g.Actual}}
	err = Col.Update(query, change)
	return err
}

func RemoveMemberFromGroup(groupId bson.ObjectId, userId bson.ObjectId) error {
	var err error
	g, err := FindGroup(groupId)
	fmt.Println("\n%s\n", userId)
	for i, oldUser := range g.UserIDs {
		if userId.Hex() == oldUser {
			g.UserIDs = append(g.UserIDs[:i], g.UserIDs[i+1:]...)
			g.Actual = append(g.Actual[:i], g.Actual[i+1:]...)
			g.Expected = append(g.Expected[:i], g.Expected[i+1:]...)
			GetGroupChanges(g)
			return err
		}
	}
	err = errors.New("Did not find member in Group")
	return err
}

func DeleteGroup(id bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("Group")
	err = Col.RemoveId(id)
	return err
}

////////////////////////////////////////////////////////
//          CONTACT FUNCTIONS         //
////////////////////////////////////////////////////////
func AddContact_other(contactName string, email string, phone string, isRealUser bool, uid bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("Contact")
	id := bson.NewObjectId()
	err = Col.Insert(&Contact{ID: id, Name: contactName, Email: email, IsRealUser: isRealUser, Timestamp: time.Now()})
	AddContactToUser(uid, id)
	return err
}

func FindContact(id bson.ObjectId) (Contact, error) {
	var err error
	Col = Session.DB("test").C("Contact")
	contact := Contact{}
	err = Col.Find(bson.M{"_id": bson.ObjectId(id)}).One(&contact)
	return contact, err
}

func GetContactChanges(c Contact) error {
	var err error
	Col = Session.DB("test").C("Contact")
	query := bson.M{"_id": c.ID}
	change := bson.M{"$set": bson.M{"name": c.Name, "phone": c.Phone, "email": c.Email, "isRealUser": c.IsRealUser}}
	err = Col.Update(query, change)
	return err
}

func DeleteContact(id bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("Contact")
	err = Col.RemoveId(id)
	return err
}

////////////////////////////////////////////////////////
//          COMMENT FUNCTIONS         //
////////////////////////////////////////////////////////

func AddComment(userName string, subject string, content string, uid bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("Comment")
	err = Col.Insert(&Comment{UserName: userName, UserID: uid.Hex(), Subject: subject, Content: content, Timestamp: time.Now()})
	return err
}

func FindCommentById(id bson.ObjectId) (Comment, error) {
	var err error
	Col = Session.DB("test").C("Comment")
	comment := Comment{}
	err = Col.Find(bson.M{"_id": bson.ObjectId(id)}).One(&comment)
	return comment, err
}

func FindCommentByUserId(id bson.ObjectId) (Comment, error) {
	var err error
	Col = Session.DB("test").C("Comment")
	comment := Comment{}
	err = Col.Find(bson.M{"userid": bson.ObjectId(id)}).One(&comment)
	return comment, err
}

func GetCommentChanges(c Comment) error {
	var err error
	Col = Session.DB("test").C("Comment")
	query := bson.M{"_id": c.ID}
	change := bson.M{"$set": bson.M{"userName": c.UserName, "userid": c.UserID, "subject": c.Subject, "content": c.Content}}
	err = Col.Update(query, change)
	return err
}

func DeleteComment(id bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("Comment")
	err = Col.RemoveId(id)
	return err
}

////////////////////////////////////////////////////////
//					PAYMENT FUNCTIONS				  //
////////////////////////////////////////////////////////

func AddPayment(payer string, payerID bson.ObjectId, payee string, payeeID bson.ObjectId, amount float32) error {
	var err error
	Col = Session.DB("test").C("Payment")
	err = Col.Insert(&Payment{Payer: payer, PayerID: payerID.Hex(), Payee: payee, PayeeID: payeeID.Hex(), Amount: amount})
	return err
}

//Only can be one payment between two people
func FindPaymentById(id bson.ObjectId) (Payment, error) {
	var err error
	Col = Session.DB("test").C("Payment")
	payment := Payment{}
	err = Col.Find(bson.M{"_id": bson.ObjectId(id)}).One(&payment)
	return payment, err
}

func FindPaymentByPayeeIdAndPayerId(payeeid bson.ObjectId, payerid bson.ObjectId) (Payment, error) {
	var err error
	Col = Session.DB("test").C("Payment")
	payment := Payment{}
	err = Col.Find(bson.M{"payeeid": payeeid.Hex(), "payerid": payerid.Hex()}).One(&payment)
	return payment, err
}

func GetPaymentChanges(p Payment) error {
	var err error
	Col = Session.DB("test").C("Payment")
	query := bson.M{"_id": p.ID}
	change := bson.M{"$set": bson.M{"payer": p.Payer, "payerid": p.PayerID, "payee": p.Payee, "payeeid": p.PayeeID, "amount": p.Amount}}
	err = Col.Update(query, change)
	return err
}

func DeletePayment(id bson.ObjectId) error {
	var err error
	Col = Session.DB("test").C("Payment")
	err = Col.RemoveId(id)
	return err
}

////////////////////////////////////////////////////////
//          TEST FUNCTIONS            //
////////////////////////////////////////////////////////

func Init() {
	ConnectToDB()
	//defer Session.Close()
	ConfigDB()

	Col = Session.DB("test").C("User")

	index := mgo.Index{
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := Col.EnsureIndex(index)
	ThisPanic(err)

}

func Close() {
	Session.Close()
}

////////////////////////////////////////////////////////
//          DATABASE FUNCTIONS          //
////////////////////////////////////////////////////////

func ConfigDB() {
	var err error
	Session.SetMode(mgo.Monotonic, true)
	// Drop Database
	if IsDrop {
		err = Session.DB("test").DropDatabase()
		ThisPanic(err)
	}
}

func ThisPanic(err error) {
	if err != nil {
		fmt.Printf("Panic: %s\n", err.Error())
	}
}

func ConnectToDB() {
	var err error
	Session, err = mgo.Dial("127.0.0.1")
	ThisPanic(err)
}
