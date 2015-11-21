//package logic
package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	//"strings"
)

type user struct {
	userName       bson.ObjectId
	expected, owed int
}

// this can be deleted
//TODO will need to work out how to divy up existing debt
func AddMember(group *[]*user, x *user) []*user {
	*group = append(*group, x)
	return *group
}

// this can deleted
//TODO work out how to spread the remainder of his money around
//MaybeTODO an error if the user is not in the group.
func RemoveMember(group *[]*user, x *user) []*user {
	t := make([]*user, len(*group)-1) // can't use the append cut trick with *s
	t = *group
	for i, ele := range *group {
		if *ele == *x {
			t = append(t[:i], t[i+1:]...)
			break
		}
	}
	//fmt.Println("t: ", t)
	*group = t
	return *group
}

func PrintGroup(group *[]*user) {
	t := make([]*user, len(*group)-1)
	t = *group
	//output := []string {""}
	for i := range *group {
		fmt.Printf(" %v ", t[i]) // use %+v for struct vals, %p for pointer
	}
	fmt.Println("")
}

// Adds a purchase for the buyer, increasing the expeced by (cost-average)
// lowers all other group members expected by average
func AddPurchase(group *[]*user, buyer *user, cost int) []*user {
	var length int = len(*group)
	var average int = (cost / length)
	for _, ele := range *group {
		if *ele == *buyer {
			ele.owed = ele.owed + cost
			ele.expected = ele.expected + (cost - average)
		} else {
			ele.expected = ele.expected - average
		}
	}
	return *group
}

// payer pays payee
// payers expected and owed increase
// payees expected and owed decrease
func PayMember(payer *user, payee *user, amount int) {
	payer.expected += amount
	payer.owed += amount
	payee.expected -= amount
	payee.owed -= amount
}

// take on the entirety of someone elses expected/owed
func TakeDebt(taker *user, giver *user) {
	taker.expected += giver.expected
	taker.owed += giver.owed
	giver.expected = 0
	giver.owed = 0
}

// split the entirety of one persons finances to other members
func SplitDebt(group *[]*user, debtHolder *user) []*user {
	var length int = (len(*group) - 1)
	for _, ele := range *group {
		if *ele != *debtHolder {
			ele.owed = ele.owed + debtHolder.owed/length
			ele.expected = ele.expected + debtHolder.expected/length
		}
	}
	debtHolder.owed = 0
	debtHolder.expected = 0
	return *group
}

func main() {
	fmt.Println("My favorite number is swag")
	K := user{"Ken", 0, 0}
	W := user{"Will", 0, 0}
	J := user{"Josh", 0, 0}
	E := user{"Evan", 0, 0}
	X := user{"Jordan", 0, 0}
	fmt.Println("Users: ", K, W, J, E)
	//Group := make([]user,0)
	Group := []*user{&K, &W, &J, &E}
	//Group = append(Group, K, W, J, E)
	//fmt.Println("Group = ", Group)
	fmt.Println("Group Creation")
	PrintGroup(&Group)

	fmt.Println("Group Add")
	AddMember(&Group, &X)
	PrintGroup(&Group)

	fmt.Println("Group Remove")
	RemoveMember(&Group, &X)
	PrintGroup(&Group) //fmt.Println("Removed Group = ", Group)

	//fmt.Println("Group Modify Values Directly")
	//Group[1].expected = 10
	//E.expected = 15
	//PrintGroup(&Group)

	fmt.Println("Add purchase for K")
	AddPurchase(&Group, &K, 90)
	PrintGroup(&Group)

	fmt.Println("W pay K 22")
	PayMember(&W, &K, 22)
	PrintGroup(&Group)

	fmt.Println("E takes Josh's debt over")
	TakeDebt(&E, &J)
	PrintGroup(&Group)

	fmt.Println("Split Ken's debt")
	SplitDebt(&Group, &K)
	PrintGroup(&Group)

	// balance check
	fmt.Println("Evan Pay Will 15")
	PayMember(&E, &W, 15)
	PrintGroup(&Group)

	fmt.Println("Evan Pay Josh 14")
	PayMember(&E, &J, 14)
	PrintGroup(&Group)

}
