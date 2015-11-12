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
	var length int = (len(*group) - 1)
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

func payMember(payer *user, payee *user, amount int) {
	payer.expected += amount
	payer.owed += amount
	payee.expected -= amount
	payee.owed -= amount
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

	fmt.Println("J pay K 5")
	payMember(&J, &K, 5)
	printGroup(&Group)
}
