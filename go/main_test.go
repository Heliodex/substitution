package main_test

import (
	"fmt"
	"testing"

	sub "github.com/Heliodex/substitution"
)

func Test(t *testing.T) {
	s := sub.Sub{
		PartNames: []sub.PartName{
			{Part: "Hello ", Name: "name"},
			{Part: "! You have ", Name: "count"},
		},
		Final: " new messages.",
	}

	fmt.Println(s)

	toSub := sub.ToSub{
		"name":  "Heliodex",
		"count": "67",
	}

	result, err := s.Sub(toSub)
	if err != nil {
		fmt.Printf("Error substituting: %v\n", err)
		return
	}

	fmt.Printf("Substituted result: %s\n", result)

	data := s.Serialise()
	fmt.Printf("Encoded data: %s\n", data)
}
