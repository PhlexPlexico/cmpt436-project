package main

import (
	"../db"
	"../logic"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

func AddPurchase_old(g db.Group, buyer string, cost int, expected []int) error {
	g = logic.AddPurchase(g, buyer, cost, expected)
	return db.GetGroupChanges(g)
}

func PayMember_old(g db.Group, payer string, payee string, amount int) error {
	g = logic.PayMember(g, payer, payee, amount)
	return db.GetGroupChanges(g)
}

func TakeDebt_old(g db.Group, taker string, payee string) error {
	g = logic.TakeDebt(g, taker, payee)
	return db.GetGroupChanges(g)
}

func main() {

	db.Init()

	_ = db.AddUser("jordan", "lys.jordan@gmail.com", "3066305775", "", true)
	_ = db.AddUser("ken", "okenso@gmail.com", "3067179886", "", true)
	_ = db.AddUser("evan", "evanclosson@gmail.com", "3067170984", "", true)
	_ = db.AddUser("Josh", "josh@usask.ca", "3067173421", "", true)
	_ = db.AddUser("William", "will@usask.ca", "3067123645", "", true)

	userid1, _ := db.FindUserIdByEmail("lys.jordan@gmail.com")
	userid2, _ := db.FindUserIdByEmail("okenso@gmail.com")
	userid3, _ := db.FindUserIdByEmail("evanclosson@gmail.com")
	userid4, _ := db.FindUserIdByEmail("josh@usask.ca")
	userid5, _ := db.FindUserIdByEmail("will@usask.ca")

	user1, _ := db.FindUserByID(userid1)
	user2, _ := db.FindUserByID(userid2)
	user3, _ := db.FindUserByID(userid3)
	user4, _ := db.FindUserByID(userid4)
	user5, _ := db.FindUserByID(userid5)

	// get user infos for sanity purpose.
	fmt.Printf("\nUser Info For User 1: %v\n", user1)
	fmt.Printf("\nUser Info For User 2: %v\n", user2)
	fmt.Printf("\nUser Info For User 3: %v\n", user3)
	fmt.Printf("\nUser Info For User 4: %v\n", user4)
	fmt.Printf("\nUser Info For User 5: %v\n", user5)

	_ = db.AddGroup("Group1", userid1)
	user1, _ = db.FindUserByID(userid1)

	// when storing an ID, use ObjectIdHex casts
	_ = db.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid2)
	_ = db.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid3)
	_ = db.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid4)
	_ = db.AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid5)

	groupid1 := user1.Groups[0]

	group1, _ := db.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nGroup Info: %v\n", group1)

	// b := [2]string{"Penn", "Teller"}

	purchase := []int{2, 2, 2, 2, 2}
	_ = AddPurchase_old(group1, userid3.Hex(), 10, purchase) // evan purchase 10
	group1, _ = db.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\n AddPurchase_old 3: %v\n", group1)

	_ = PayMember_old(group1, userid1.Hex(), userid3.Hex(), 2) // Jordan pays Evan 2
	group1, _ = db.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nPayMember_old 1, 3: %v\n", group1)

	_ = TakeDebt_old(group1, userid2.Hex(), userid3.Hex()) // ken taking evans
	group1, _ = db.FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nTake Debt 2, 3: %v\n", group1)

	db.Close()
	// user = db.FindUser("email")

	// AddGroup("groupa" GetIDbyEmail("email"))

	// group = db.FindGroup("groupID")
	// fmt.Printf(" %v ", t[i]) // use %+v for struct vals, %p for pointer

	// //logic.AddPurchase_old(group db.FindUser("email") cost)

}
