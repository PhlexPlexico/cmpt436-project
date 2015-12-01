package main

import (
	// "../logic"
	"../server"
	//"fmt"
	"gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
)

var (
	IsDrop  = true
	Session *mgo.Session
	Col     *mgo.Collection
)

func main() {
	Session, Col, IsDrop = server.Init()
	// var err error
	_ = server.AddUser("jordan", "asdf@mail.com", "123", true)

	// user = server.FindUser("email")

	// AddGroup("groupa" GetIDbyEmail("email"))

	// group = server.FindGroup("groupID")
	// fmt.Printf(" %v ", t[i]) // use %+v for struct vals, %p for pointer

	// //logic.AddPurchase(group server.FindUser("email") cost)

}
