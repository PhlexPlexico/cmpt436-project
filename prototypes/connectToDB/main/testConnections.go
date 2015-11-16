package main

import (
 	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"fmt"
)


type User struct {
	ID        	bson.ObjectId 	`json:"id" bson:"_id,omitempty"`
	Name      	string			`json:"name"`
	Phone     	string			`json:"phone"`
	Email		string			`json:"email"`
	IsRealUser	bool			`json:"isRealUser"`
	Groups		[]*Group 		`json:"groups"`
	Contacts	[]*Contact		`json:"contacts"`
	Timestamp 	time.Time 		`json:"time" bson:"time.Time"`
}

type Contact struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id,omitempty"`
	Name		string 			`json:"name"`
	Phone		string  		`json:"phone"`
	Email		string 			`json:"email"`
	IsRealUser	bool			`json:"isRealUser`
	Timestamp 	time.Time 		`json:"time" bson:"time.Time"`
}

type Group struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id, omitempty"`
	GroupName	string 			`json:"groupName"`
	Users 		[]*User			`json:"users"`
	Expected	[]*int 			`json:"expected"`
	Actual		[]*int 			`json:"actual"`
}

type Comment struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id, omitempty"`
	UserName	string 			`json:"userName"`
	Subject		string 			`json:"subject"`
	Content		string 			`json:"content"`
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

	err = c.Insert(&User{ Name: "Ale", Phone: "+922", IsRealUser: true, Email:"abc@gmail.com", Timestamp: time.Now()})
	err = c.Insert(&User{ Name: "Jrock", Phone: "+911", IsRealUser: true, Email:"jcl@gmail.com", Timestamp: time.Now()})	
	if err != nil {
		panic(err)
	}

	c = session.DB("test").C("Contact")
	err = c.Insert(&Contact{ Name: "Ale", Phone: "+922", IsRealUser: true, Email:"abc@gmail.com", Timestamp: time.Now()})



	result := Contact{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n")
	fmt.Println(result)
	fmt.Println("\n")

	findJ := User{}
	c = session.DB("test").C("User")
	err = c.Find(bson.M{"name": "Jrock"}).Select(bson.M{"_id":1}).One(&findJ)
	fmt.Println(findJ)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n")
	fmt.Println(findJ.ID.Hex())
	fmt.Println("\n")

	hex := findJ.ID.Hex()

	// query := bson.M{"_id": bson.ObjectId(hex)}
	// change := bson.M{"$push": bson.M{"Contacts": &result}}
	// fmt.Println(change)
	// err = c.Update(query, change)

	//  findJ = User{}
	//  err = c.Find(bson.M{"name": "Jrock"}).One(&findJ)
	//  if err != nil {
	//   	panic(err)
	//  }

	//  fmt.Println(findJ.Contacts)






}	

