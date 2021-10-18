package main

import (
	"sort"
	"fmt"
)

type student struct {

	FirstName string
	LastName string
}

// sortListOfStudents sorts a list of student per lastname first
// and per firstname in worse case.
func sortListOfStudents(students []student) {
	
	sort.Slice(students, func(i, j int) bool { 
		if students[i].LastName != students[j].LastName {
			return students[i].LastName < students[j].LastName
		}

		return students[i].FirstName <= students[j].FirstName
	})


}

func main() {
	
	students := []student{
		{FirstName:"jerome1", LastName:"amon"}, // should be 2nd
		{FirstName:"jerome0", LastName:"amon"}, // should be 1st
		{FirstName:"jerome", LastName:"amon2"}, // should be 5th
		{FirstName:"jerome", LastName:"amon1"}, // should be 4th
		{FirstName:"jerome", LastName:"amon0"}, // should be 3rd
	}
	// uncomment to view initial order 
	// fmt.Println(students)
	sortListOfStudents(students)
	fmt.Println(students)
}

/* output:
[11:45:09] {nxos-geek}:~$ go run code-example.go
[{jerome0 amon} {jerome1 amon} {jerome amon0} {jerome amon1} {jerome amon2}]
*/