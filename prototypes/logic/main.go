package main

import (
	"fmt"
	//"strings"
)

type user struct {
	userName string
	expected, owed int
}


//TODO will need to work out how to divy up existing debt
func addMember(s *[]*user, x *user) []*user {
	*s = append(*s, x)
	return *s
}


//TODO work out how to spread the remainder of his money around
//MaybeTODO an error if the user is not in the group.
func removeMember(s *[]*user, x *user) []*user {
	t := make([]*user, len(*s)-1) // can't use the append cut trick with *s
	t = *s
	for i,ele := range *s {
  		if *ele == *x {
  			t = append(t[:i], t[i+1:]...)
  			break
  		}
	}
	//fmt.Println("t: ", t)
	*s = t
	return *s
}

func printGroup(s *[]*user) {
	t := make([]*user, len(*s)-1)
	t = *s
 	//output := []string {""}
		for i := range *s {
			fmt.Printf(" %v ", t[i]) // use %+v for struct vals, %p for pointer
 	}
 	fmt.Println("")
}


func main() {
	fmt.Println("My favorite number is swag")
	K := user{"Ken", 0, 0}
	W := user{"Will", 0, 0}
	J := user{"Josh", 0, 0}
	E := user{"Evan", 0 ,0}
	X := user{"Jordan", 0 ,0}
	fmt.Println("Users: ", K, W, J, E)
	//Group := make([]user,0)
	Group := []*user {&K, &W, &J, &E}
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

	fmt.Println("Group Modify Values Directly")
	Group[1].expected = 10
	E.expected = 15
	printGroup(&Group)

}

