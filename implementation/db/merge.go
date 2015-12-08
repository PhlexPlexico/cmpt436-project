package db

import (
	"../logic"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

func AddPurchase(g *Group, buyer string, cost int, expected []int) error {
	logic.AddPurchase(g, buyer, cost, expected)
	Purchase := &Purchase{
		PayerID:       buyer,
		AmountInCents: cost,
		Expected:      expected,
	}
	PurchaseBytes, err := json.Marshal(Purchase)
	if err != nil {
		return err
	}
	PurchseFeedItem := &FeedItem{
		Content: PurchaseBytes,
		GroupID: g.ID,
		Type:    FeedItemTypePurchase,
	}
	AddFeedItemToGroupByID(g, PurchaseFeedItem)
	return GetGroupChanges(g)
}

func PayMember(g *Group, payer string, payee string, amount int) error {
	logic.PayMember(g, payer, payee, amount)
	//yolo := bson.NewObjectId()
	Payment := &db.Payment{
		//ID:            yolo,
		PayerID:       payer,
		PayeeID:       payee,
		AmountInCents: amount,
	}
	PaymentBytes, err := json.Marshal(Payment)
	if err != nil {
		return err
	}
	PaymentFeedItem := &db.FeedItem{
		Content: PaymentBytes,
		GroupID: g.ID.Hex(),
		Type:    db.FeedItemTypePayment,
	}
	log.Printf("\n\n PaymentFeedItem %v \n \n \n", PaymentFeedItem)
	db.AddFeedItemToGroup(g, PaymentFeedItem)
	return db.GetGroupChanges(g)
}

func TakeDebt(g *Group, taker string, payee string) error {
	logic.TakeDebt(g, taker, payee)
	return GetGroupChanges(g)
}

func main() {

	Init()

	_ = AddUser("jordan", "lys.jordan@gmail.com", "3066305775", "", true)
	_ = AddUser("ken", "okenso@gmail.com", "3067179886", "", true)
	_ = AddUser("evan", "evanclosson@gmail.com", "3067170984", "", true)
	_ = AddUser("Josh", "josh@usask.ca", "3067173421", "", true)
	_ = AddUser("William", "will@usask.ca", "3067123645", "", true)

	userid1, _ := FindUserIdByEmail("lys.jordan@gmail.com")
	userid2, _ := FindUserIdByEmail("okenso@gmail.com")
	userid3, _ := FindUserIdByEmail("evanclosson@gmail.com")
	userid4, _ := FindUserIdByEmail("josh@usask.ca")
	userid5, _ := FindUserIdByEmail("will@usask.ca")

	user1, _ := FindUserByID(userid1)
	user2, _ := FindUserByID(userid2)
	user3, _ := FindUserByID(userid3)
	user4, _ := FindUserByID(userid4)
	user5, _ := FindUserByID(userid5)

	// get user infos for sanity purpose.
	fmt.Printf("\nUser Info For User 1: %v\n", user1)
	fmt.Printf("\nUser Info For User 2: %v\n", user2)
	fmt.Printf("\nUser Info For User 3: %v\n", user3)
	fmt.Printf("\nUser Info For User 4: %v\n", user4)
	fmt.Printf("\nUser Info For User 5: %v\n", user5)

	_ = AddGroup("Group1", userid1)
	user1, _ = FindUserByID(userid1)

	// when storing an ID, use ObjectIdHex casts
	_ = AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid2)
	_ = AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid3)
	_ = AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid4)
	_ = AddMemberToGroupByID(bson.ObjectIdHex(user1.Groups[0]), userid5)

	groupid1 := user1.Groups[0]

	group1, _ := FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nGroup Info: %v\n", group1)

	// b := [2]string{"Penn", "Teller"}

	purchase := []int{2, 2, 2, 2, 2}
	_ = AddPurchase(group1, userid3.Hex(), 10, purchase) // evan purchase 10
	group1, _ = FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\n AddPurchase 3: %v\n", group1)

	_ = PayMember(group1, userid1.Hex(), userid3.Hex(), 2) // Jordan pays Evan 2
	group1, _ = FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nPayMember 1, 3: %v\n", group1)

	_ = TakeDebt(group1, userid2.Hex(), userid3.Hex()) // ken taking evans
	group1, _ = FindGroup(bson.ObjectIdHex(groupid1))
	fmt.Printf("\nTake Debt 2, 3: %v\n", group1)

	Close()
	// user = FindUser("email")

	// AddGroup("groupa" GetIDbyEmail("email"))

	// group = FindGroup("groupID")
	// fmt.Printf(" %v ", t[i]) // use %+v for struct vals, %p for pointer

	// //logic.AddPurchase(group FindUser("email") cost)

}
