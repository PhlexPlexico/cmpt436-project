package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"sort"
	"time"
)

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

type Contact struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	IsRealUser bool          `json:"isRealUser`
	Timestamp  time.Time     `json:"time"`
}

type Group struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	GroupName string        `json:"groupName"`
	UserIDs   []string      `json:"userids"`
	Expected  []float32     `json:"expected"`
	Actual    []float32     `json:"actual"`
}

type Comment struct {
	ID        bson.ObjectId `json:"id" bson:"_id, omitempty"`
	UserName  string        `json:"userName"`
	Subject   string        `json:"subject"`
	Content   string        `json:"content"`
	Timestamp time.Time     `json:"time"`
}

type Payment struct {
	ID            bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Payer         string        `json:"payer"`
	Payee         string        `json:"payee"`
	AmountInCents int           `json:"amountInCents"`
	Timestamp     time.Time     `json:"time"`
}

type Purchase struct {
	ID            bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Payer         string        `json:"payer"`
	AmountInCents int           `json:"amountInCents"`
	Timestamp     time.Time     `json:"time"`
}

type Notification struct {
	ID        bson.ObjectId `json:"id" bson:"_id, omitempty"`
	Subject   string        `json:"subject"`
	Content   string        `json:"content"`
	Timestamp time.Time     `json:"time"`
}

var (
	IsDrop  = true
	Session *mgo.Session
	Col     *mgo.Collection
	err     error
)

///////////////////////////////////////////////////////////

func AddUser(name string, email string, phone string, isRealUser bool) {
	Col = Session.DB("test").C("User")
	err = Col.Insert(&User{Name: name, Phone: phone, IsRealUser: isRealUser, Email: email, Timestamp: time.Now()})
	ThisPanic(err)
}

func FindUserByID(id bson.ObjectId) User {
	Col = Session.DB("test").C("User")
	user := User{}
	err = Col.Find(bson.M{"_id": bson.ObjectId(id)}).One(&user)
	ThisPanic(err)
	return user
}

func FindUserIdByEmail(email string) bson.ObjectId {
	Col = Session.DB("test").C("User")
	user := User{}
	err = Col.Find(bson.M{"email": email}).One(&user)
	ThisPanic(err)
	return user.ID
}

func AddGroupToUser(userId bson.ObjectId, groupId bson.ObjectId) {
	Col = Session.DB("test").C("User")
	query := bson.M{"_id": bson.ObjectId(userId)}
	change := bson.M{"$push": bson.M{"groups": groupId.Hex()}}
	err = Col.Update(query, change)
	ThisPanic(err)
}

////////////////////////////////////////////////////////

func AddGroup(groupName string, uid bson.ObjectId) bool {

	Col = Session.DB("test").C("Group")
	id := bson.NewObjectId()
	err = Col.Insert(&Group{ID: id, GroupName: groupName, UserIDs: []string{uid.Hex()}, Expected: []float32{0}, Actual: []float32{0}})
	ThisPanic(err)
	AddGroupToUser(uid, id)
	return true
}

func FindGroup(id bson.ObjectId) Group {
	Col = Session.DB("test").C("Group")
	group := Group{}
	err = Col.Find(bson.M{"_id": bson.ObjectId(id)}).One(&group)
	ThisPanic(err)
	return group
}

func AddMemberToGroupByID(groupId bson.ObjectId, userId bson.ObjectId) bool {
	g := FindGroup(groupId)
	Col = Session.DB("test").C("Group")
	query := bson.M{"_id": g.ID}
	change := bson.M{"$push": bson.M{"userids": userId.Hex(), "expected": 0, "actual": 0}}
	err = Col.Update(query, change)
	return true
}

func GetGroupChanges(g Group) {
	Col = Session.DB("test").C("Group")
	query := bson.M{"_id": g.ID}
	change := bson.M{"$set": bson.M{"groupName": g.GroupName, "users": g.UserIDs, "expected": g.Expected, "actual": g.Actual}}
	err = Col.Update(query, change)
	ThisPanic(err)
}

func RemoveMemberFromGroup(groupId bson.ObjectId, userId bson.ObjectId) bool {
	g := FindGroup(groupId)
	fmt.Println("\n%s\n", userId)
	for i, oldUser := range g.UserIDs {
		if userId.Hex() == oldUser {
			g.UserIDs = append(g.UserIDs[:i], g.UserIDs[i+1:]...)
			g.Actual = append(g.Actual[:i], g.Actual[i+1:]...)
			g.Expected = append(g.Expected[:i], g.Expected[i+1:]...)
			GetGroupChanges(g)
			return true
		}
	}
	return false
}

func DeleteGroup(id bson.ObjectId) bool {
	Col = Session.DB("test").C("Group")
	err = Col.RemoveId(id)
	ThisPanic(err)
	return true
}

////////////////////////////////////////////////////////

func main() {

	ConnectToDB()
	defer Session.Close()
	ConfigDB()

	Col = Session.DB("test").C("User")

	index := mgo.Index{
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = Col.EnsureIndex(index)

	ThisPanic(err)

	// test Functions for Users

	// add Users to DB
	AddUser("blah", "abc@mail.com", "12334", true)
	AddUser("jrock", "asdf@mail.com", "12345", true)
	AddUser("plexico", "bvcx@mail.com", "12321", true)
	AddUser("garmu", "zcxv@mail.com", "12314", true)

	id1 := FindUserIdByEmail("abc@mail.com")
	id2 := FindUserIdByEmail("asdf@mail.com")
	id3 := FindUserIdByEmail("bvcx@mail.com")
	id4 := FindUserIdByEmail("zcxv@mail.com")

	fmt.Printf("\nUserId1: %s\n", id1)

	//Add Users to Groups
	AddGroup("group1", id1)

	user1 := FindUserByID(id1)

	groupid1 := bson.ObjectIdHex(user1.Groups[0])
	fmt.Printf("Group1: %s\n", groupid1)

	AddMemberToGroupByID(groupid1, id2)
	AddMemberToGroupByID(groupid1, id3)
	AddMemberToGroupByID(groupid1, id4)

	group1 := FindGroup(groupid1)
	fmt.Printf("Group1: %s\n", group1.UserIDs)

	RemoveMemberFromGroup(groupid1, id2)

}

func ConfigDB() {
	Session.SetMode(mgo.Monotonic, true)
	// Drop Database
	if IsDrop {
		err = Session.DB("test").DropDatabase()
		ThisPanic(err)
	}
}

func ThisPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func ConnectToDB() {
	Session, err = mgo.Dial("127.0.0.1")
	ThisPanic(err)
}
