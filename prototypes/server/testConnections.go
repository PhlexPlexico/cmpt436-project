package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	IsRealUser bool          `json:"isRealUser"`
	Groups     []Group      `json:"groups" bson:"groups"`
	Contacts   []Contact    `json:"contacts" bson:"contacts"`
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
	ID        bson.ObjectId 	`json:"id" bson:"_id,omitempty"`
	GroupName string        	`json:"groupName" bson:"groupName"`
	UserIDs   []string			`json:"users" bson:"users"`
	Expected  []int        		`json:"expected"`
	Actual    []int        		`json:"actual"`
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
	IsDrop     	= true
	Session    	*mgo.Session
	Col 		*mgo.Collection
	err 		error
)

func ThisPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func ConnectToDB() {
	Session, err = mgo.Dial("127.0.0.1")
	ThisPanic(err)
}


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

func GetIDbyEmail(email string) string {
	Col = Session.DB("test").C("User")
	user := User{}
	err = Col.Find(bson.M{"email": email}).One(&user)
	ThisPanic(err)
	return user.ID.Hex()
}

func AddGroup(groupName string, uid bson.ObjectId) bool {
	Col = Session.DB("test").C("Group")
	id := bson.NewObjectId()
	err = Col.Insert(&Group{ID: id, GroupName: groupName, UserIDs: []string{uid.Hex()}})
	ThisPanic(err)
	actualGroup := Group{}
	err = Col.Find(bson.M{"_id": id}).All(&actualGroup)
	ThisPanic(err)
	query := bson.M{"_id": id}
	Col = Session.DB("test").C("User")
	change := bson.M{"$push": bson.M{"groups": actualGroup}}
	err = Col.Update(query, change)
	ThisPanic(err)
	return true
}

func FindGroup(id bson.ObjectId) Group {
	Col = Session.DB("test").C("Group")
	actualGroup := Group{}
	err = Col.Find(bson.M{"_id": id}).All(&actualGroup)
	ThisPanic(err)
	return actualGroup
}

func AddMemberToGroupByID(groupId bson.ObjectId, userId bson.ObjectId ) bool {
	foundGroup := FindGroup(groupId)
	t := AddGroup(foundGroup.GroupName, userId)
	return t

}
	
func GetGroupChanges(g Group) {
	Col = Session.DB("test").C("Group")
	query := bson.M{"_id": g.ID}
	change := bson.M{"$push": bson.M{"_id": g.ID, "groupName": g.GroupName, "users": g.UserIDs, "expected": g.Expected, "actual": g.Actual}}
	err = Col.Update(query, change)
	ThisPanic(err)
}

// func RemoveMemberFromGroup(groupId bson.ObjectId, userId bson.ObjectId ) {
// 	group := FindGroup(groupId)
	
	


	
// }

func DeleteGroup(id bson.ObjectId) bool {
	Col = Session.DB("test").C("Group")
	err = Col.RemoveId(id)
	ThisPanic(err)
	return true
}


func ConfigDB() {
	Session.SetMode(mgo.Monotonic, true)
	// Drop Database
	if IsDrop {
		err = Session.DB("test").DropDatabase()
		ThisPanic(err)
	}
}

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


	err = Col.Insert(&User{Name: "Ale", Phone: "+922", IsRealUser: true, Email: "abc@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)
	err = Col.Insert(&User{Name: "Jrock", Phone: "+911", IsRealUser: true, Email: "jcl@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)

	Col = Session.DB("test").C("Contact")
	err = Col.Insert(&Contact{Name: "Ale", Phone: "+922", IsRealUser: true, Email: "abc@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)

	result := Contact{}
	err = Col.Find(bson.M{"name": "Ale"}).One(&result)
	ThisPanic(err)

	fmt.Println("\n")
	fmt.Println(result)
	fmt.Println("\n")

	findJ := User{}
	Col = Session.DB("test").C("User")
	err = Col.Find(bson.M{"name": "Jrock"}).Select(bson.M{"_id": 1}).One(&findJ)
	fmt.Println(findJ)
	ThisPanic(err)

	fmt.Println("\nHexID of JRock\n")
	fmt.Println(findJ.ID.Hex())
	fmt.Println("\nResult Object\n")
	fmt.Println(result)

	query := bson.M{"_id": bson.ObjectId(findJ.ID)}
	fmt.Println("\nQuery\n")
	fmt.Println(query)
	change := bson.M{"$push": bson.M{"contacts": result}}
	//change2 := bson.M{"$push": bson.M{"contacts": bson.M{"name": result.Name}}}

	fmt.Println("\nUpdate Params\n")
	fmt.Println(change)
	err = Col.Update(query, change)
	ThisPanic(err)

	findJ = User{}
	err = Col.Find(bson.M{"name": "Jrock"}).One(&findJ)
	ThisPanic(err)

	fmt.Println("\nContacts of JRock\n")
	fmt.Println(findJ.Contacts[0])

	Col = Session.DB("test").C("Contact")
	err = Col.Insert(&Contact{Name: "Eclo", Phone: "+306", IsRealUser: true, Email: "eclo@gmail.com", Timestamp: time.Now()})
	ThisPanic(err)
	result = Contact{}
	err = Col.Find(bson.M{"name": "Eclo"}).One(&result)

	Col = Session.DB("test").C("User")
	/*ADD ANOTHER CONTACT*/
	findJ = User{}
	Col = Session.DB("test").C("User")
	err = Col.Find(bson.M{"name": "Jrock"}).Select(bson.M{"_id": 1}).One(&findJ)
	fmt.Println(findJ)
	
	ThisPanic(err)

	fmt.Println("\nHexID of JRock\n")
	fmt.Println(findJ.ID.Hex())
	fmt.Println("\nResult Object\n")
	fmt.Println(result)

	query = bson.M{"_id": bson.ObjectId(findJ.ID)}
	fmt.Println("\nQuery\n")
	fmt.Println(query)
	change = bson.M{"$push": bson.M{"contacts": result}}
	//change2 := bson.M{"$push": bson.M{"contacts": bson.M{"name": result.Name}}}

	fmt.Println("\nUpdate Params\n")
	fmt.Println(change)
	err = Col.Update(query, change)
	ThisPanic(err)

	value := findJ.ID.Hex()
	array := []string{value}
	fmt.Println(array[0])
	Col = Session.DB("test").C("Group")
	err = Col.Insert(&Group{GroupName: "test", UserIDs: array})
	ThisPanic(err)


	g := Group{}

	err = Col.Find(bson.M{"groupName": "test"}).Select(bson.M{"_id": 1}).One(&g)
	ThisPanic(err)
	fmt.Println("\n")
	fmt.Printf("%v",g)
	fmt.Println("\n")
}
