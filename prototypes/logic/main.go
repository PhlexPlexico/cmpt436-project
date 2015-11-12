package main

import (
	"fmt"
	//"strings"
)

type user struct {
	userName       string
	expected, owed int
}

//TODO will need to work out how to divy up existing debt
func addMember(group *[]*user, x *user) []*user {
	*group = append(*group, x)
	return *group
}

//TODO work out how to spread the remainder of his money around
//MaybeTODO an error if the user is not in the group.
func removeMember(group *[]*user, x *user) []*user {
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

func printGroup(group *[]*user) {
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
func addPurchase(group *[]*user, buyer *user, cost int) []*user {
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
func payMember(payer *user, payee *user, amount int) {
	payer.expected += amount
	payer.owed += amount
	payee.expected -= amount
	payee.owed -= amount
}

// take on the entirety of someone elses expected/owed
func takeDebt(taker *user, giver *user) {
	taker.expected += giver.expected
	taker.owed += giver.owed
	giver.expected = 0
	giver.owed = 0
}

// split the entirety of one persons finances to other members
func splitDebt(group *[]*user, debtHolder *user) []*user {
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
	printGroup(&Group)

	fmt.Println("Group Add")
	addMember(&Group, &X)
	printGroup(&Group)

	fmt.Println("Group Remove")
	removeMember(&Group, &X)
	printGroup(&Group) //fmt.Println("Removed Group = ", Group)

	//fmt.Println("Group Modify Values Directly")
	//Group[1].expected = 10
	//E.expected = 15
	//printGroup(&Group)

	fmt.Println("Add purchase for K")
	addPurchase(&Group, &K, 90)
	printGroup(&Group)

	fmt.Println("W pay K 22")
	payMember(&W, &K, 22)
	printGroup(&Group)

	fmt.Println("E takes Josh's debt over")
	takeDebt(&E, &J)
	printGroup(&Group)

	fmt.Println("Split Ken's debt")
	splitDebt(&Group, &K)
	printGroup(&Group)

	// balance check
	fmt.Println("Evan Pay Will 15")
	payMember(&E, &W, 15)
	printGroup(&Group)

	fmt.Println("Evan Pay Josh 14")
	payMember(&E, &J, 14)
	printGroup(&Group)

}
