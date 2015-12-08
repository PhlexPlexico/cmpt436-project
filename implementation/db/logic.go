//package logic
package db

import (
	"fmt"
	//"gopkg.in/mgo.v2/bson"
	//"strings"
	//"sort"
)

// type user struct {
// 	userName       bson.ObjectId
// 	Actual, Expected int
// }

// this can be deleted
//TODO will need to work out how to divy up existing debt
// func AddMember(group *[]*user, x *user) []*user {
// 	*group = append(*group, x)
// 	return *group
// }

// // this can deleted
// //TODO work out how to spread the remainder of his money around
// //MaybeTODO an error if the user is not in the group.
// func RemoveMember(group *[]*user, x *user) []*user {
// 	t := make([]*user, len(*group)-1) // can't use the append cut trick with *s
// 	t = *group
// 	for i, ele := range *group {
// 		if *ele == *x {
// 			t = append(t[:i], t[i+1:]...)
// 			break
// 		}
// 	}
// 	//fmt.Println("t: ", t)
// 	*group = t
// 	return *group
// }

// func PrintGroup(group *[]*user) {
// 	t := make([]*user, len(*group)-1)
// 	t = *group
// 	//output := []string {""}
// 	for i := range *group {
// 		fmt.Printf(" %v ", t[i]) // use %+v for struct vals, %p for pointer
// 	}
// 	fmt.Println("")
// }

// Adds a purchase for the buyer, increasing the expeced by (cost-average)
// lowers all other group members Actual by average
func ProcessPurchase(group *Group, buyer string, cost int, expected []int) {
	//var length int = len(group.UserIDs)
	//var average int = (expected / length)
	for ele := range group.UserIDs {
		if group.UserIDs[ele] == buyer {
			group.Expected[ele] = group.Expected[ele] + expected[ele]
			group.Actual[ele] = group.Actual[ele] + cost
		} else {
			group.Expected[ele] = group.Actual[ele] + expected[ele]
		}
	}
}

// payer pays payee
// payers Actual and Expected increase
// payees Actual and Expected decrease
func ProcessPayment(group *Group, payer string, payee string, amount int) {
	payerPos, payeePos := getPositions(group, payer, payee)

	fmt.Printf("\n Begin payer: %v %v", group.Expected[payerPos], group.Actual[payerPos])
	//fmt.Printf("\n Begin payee: %v %v",group.Expected[payeePos], group.Actual[payeePos] )

	group.Actual[payerPos] += amount
	group.Expected[payerPos] -= amount
	group.Actual[payeePos] -= amount
	group.Expected[payeePos] += amount

	//fmt.Printf("\n End payer: %v %v",group.Expected[payerPos], group.Actual[payerPos] )
	//fmt.Printf("\n End payee: %v %v",group.Expected[payeePos], group.Actual[payeePos] )
}

func getPositions(group *Group, u1 string, u2 string) (int, int) {
	x := -1
	y := -1
	for ele := range group.UserIDs {
		if group.UserIDs[ele] == u1 {
			x = ele
		} else if group.UserIDs[ele] == u2 {
			y = ele
		}
	}
	return x, y
}

// take on the entirety of someone elses Actual/Expected
func ProcessTakeDebt(group *Group, taker string, giver string) {
	takerPos, giverPos := getPositions(group, taker, giver)
	group.Actual[takerPos] += group.Actual[giverPos]
	group.Expected[takerPos] += group.Expected[giverPos]
	group.Actual[giverPos] = 0
	group.Expected[giverPos] = 0
}

// this will be automatic with the sliders defualt when take debt uses default
// DELETE
// split the entirety of one persons finances to other members
// func SplitDebt(group *[]*user, debtHolder *user) []*user {
// 	var length int = (len(*group) - 1)
// 	for _, ele := range *group {
// 		if *ele != *debtHolder {
// 			ele.Expected = ele.Expected + debtHolder.Expected/length
// 			ele.Actual = ele.Actual + debtHolder.Actual/length
// 		}
// 	}
// 	debtHolder.Expected = 0
// 	debtHolder.Actual = 0
// 	return *group
// }

// func main() {
// 	fmt.Println("My favorite number is swag")
// 	K := user{"Ken", 0, 0}
// 	W := user{"Will", 0, 0}
// 	J := user{"Josh", 0, 0}
// 	E := user{"Evan", 0, 0}
// 	X := user{"Jordan", 0, 0}
// 	fmt.Println("Users: ", K, W, J, E)
// 	//Group := make([]user,0)
// 	Group := []*user{&K, &W, &J, &E}
// 	//Group = append(Group, K, W, J, E)
// 	//fmt.Println("Group = ", Group)
// 	fmt.Println("Group Creation")
// 	PrintGroup(&Group)

// 	fmt.Println("Group Add")
// 	AddMember(&Group, &X)
// 	PrintGroup(&Group)

// 	fmt.Println("Group Remove")
// 	RemoveMember(&Group, &X)
// 	PrintGroup(&Group) //fmt.Println("Removed Group = ", Group)

// 	//fmt.Println("Group Modify Values Directly")
// 	//Group[1].Actual = 10
// 	//E.Actual = 15
// 	//PrintGroup(&Group)

// 	fmt.Println("Add purchase for K")
// 	AddPurchase(&Group, &K, 90)
// 	PrintGroup(&Group)

// 	fmt.Println("W pay K 22")
// 	PayMember(&W, &K, 22)
// 	PrintGroup(&Group)

// 	fmt.Println("E takes Josh's debt over")
// 	TakeDebt(&E, &J)
// 	PrintGroup(&Group)

// 	fmt.Println("Split Ken's debt")
// 	SplitDebt(&Group, &K)
// 	PrintGroup(&Group)

// 	// balance check
// 	fmt.Println("Evan Pay Will 15")
// 	PayMember(&E, &W, 15)
// 	PrintGroup(&Group)

// 	fmt.Println("Evan Pay Josh 14")
// 	PayMember(&E, &J, 14)
// 	PrintGroup(&Group)

// }
