package main

import (
	"../logic"
	"../server"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

func AddPurchase(g server.Group, buyer string, cost int, expected []int) error {
	g = logic.AddPurchase(g, buyer, cost, expected)
	return server.GetGroupChanges(g)
}

func PayMember(g server.Group, payer string, payee string, amount int) error {
	g = logic.PayMember(g, payer, payee, amount)
	return server.GetGroupChanges(g)
}

func TakeDebt(g server.Group, taker string, payee string) error {
	g = logic.TakeDebt(g, taker, payee)
	return server.GetGroupChanges(g)
}

func main() {

	server.Init()

	_ = server.AddUser("jordan", "lys.jordan@gmail.com", "3066305775", true)
	_ = server.AddUser("ken", "okenso@gmail.com", "3067179886", true)
	_ = server.AddUser("evan", "evanclosson@gmail.com", "3067170984", true)
	_ = server.AddUser("Josh", "josh@usask.ca", "3067173421", true)
	_ = server.AddUser("William", "will@usask.ca", "3067123645", true)

	userid1, _ := server.FindUserIdByEmail("lys.jordan@gmail.com")
	userid2, _ := server.FindUserIdByEmail("okenso@gmail.com")
	userid3, _ := server.FindUserIdByEmail("evanclosson@gmail.com")
	userid4, _ := server.FindUserIdByEmail("josh@usask.ca")
	userid5, _ := server.FindUserIdByEmail("will@usask.ca")

	user1, _ := server.FindUserByID(userid1)
	user2, _ := server.FindUserByID(userid2)
	user3, _ := server.FindUserByID(userid3)
	user4, _ := server.FindUserByID(userid4)
	user5, _ := server.FindUserByID(userid5)

	// get user infos for sanity purpose.
	fmt.Printf("\nUser Info For User 1: %v\n", user1)
	fmt.Printf("\nUser Info For User 2: %v\n", user2)
	fmt.Printf("\nUser Info For User 3: %v\n", user3)
	fmt.Printf("\nUser Info For User 4: %v\n", user4)
	fmt.Printf("\nUser Info For User 5: %v\n", user5)

	_ = server.AddGroup("Group1", userid1)
	user1, _ = server.FindUserByID(userid1)

	// when storing an ID, use ObjectIdHex casts
	_ = server.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid2)
	_ = server.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid3)
	_ = server.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid4)
	_ = server.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid5)

	groupid1 := user1.Groups[0]

	group1, _ := server.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nGroup Info: %v\n", group1)

	// b := [2]string{"Penn", "Teller"}

	purchase := []int{2, 2, 2, 2, 2}
	_ = AddPurchase(group1, userid3.Hex(), 10, purchase) // evan purchase 10
	group1, _ = server.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\n AddPurchase 3: %v\n", group1)

	_ = PayMember(group1, userid1.Hex(), userid3.Hex(), 2) // Jordan pays Evan 2
	group1, _ = server.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nPayMember 1, 3: %v\n", group1)

	_ = TakeDebt(group1, userid2.Hex(), userid3.Hex()) // ken taking evans
	group1, _ = server.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nTake Debt 2, 3: %v\n", group1)

	server.Close()
	// user = server.FindUser("email")

	// AddGroup("groupa" GetIDbyEmail("email"))

	// group = server.FindGroup("groupID")
	// fmt.Printf(" %v ", t[i]) // use %+v for struct vals, %p for pointer

	// //logic.AddPurchase(group server.FindUser("email") cost)

}
