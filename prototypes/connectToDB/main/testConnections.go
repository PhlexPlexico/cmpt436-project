package main

import (
 	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)


type User struct {
	ID        	bson.ObjectId 	`json:"id" bson:"_id,omitempty"`
	Name      	string			`json:"name" bson:"name"`
	Phone     	string			`json:"phone" bson:"phone"`
	Email		string			`json:"email" bson:"email"`
	IsRealUser	bool			`json:"isRealUser" bson:"isRealUser"`
	Groups		[]*Group 		`json:"groups" bson:"groups"`
	Contacts	[]*Contact		`json:"contacts" bson:"contacts"`
	Timestamp 	time.Time 		`json:"time" bson:"time.Time"`
}

type Contact struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id"`
	Name		string 			`json:"name" bson:"name"`
	Phone		string  		`json:"phone" bson:"phone"`
	Email		string 			`json:"email" bson:"email"`

}

type Group struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id"`
	GroupName	string 			`json:"groupName" bson:"groupName"`
	Users 		[]User			`json:"users" bson:"users"`
}

type Comment struct {
	ID 			bson.ObjectId 	`bson:"_id"`
	UserName	string 			`json:"userName" bson:"userName"`
	Subject		string 			`json:"subject" bson:"subject"`
	Content		string 			`json:"content" bson:"content"`
	Timestamp	time.Time  		`json:"time" bson:"time.Time"`
}

var (
	IsDrop = true
)

func main() {

	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	// Drop Database
	if IsDrop {
		err = session.DB("test").DropDatabase()
		if err != nil {
			panic(err)
		}
	}
	c := session.DB("test").C("User")


	index := mgo.Index {
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)

	if err != nil {
		panic(err)
	}

	err = c.Insert(&User{ Name: "Ale", Phone: "+55 53 1234 4321", IsRealUser: true, Email:"abc@gmail.com", Timestamp: time.Now()})


	
	if err != nil {
		panic(err)
	}


}