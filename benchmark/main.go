package main

import (
	"github.com/lologarithm/netgen/benchmark/netmsg"
)

func main() {
	m := &netmsg.Benchy{
		Name:     "asdfasdfasdfasdf",
		BirthDay: 1234567801,
		Phone:    "123-456-7890",
		Siblings: 66,
		Spouse:   1,
		Money:    123.456,
	}
	print(m)
	// stub so gopherjs stops whining
}
