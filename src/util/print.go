package util

import (
	"encoding/json"
	"fmt"
)

func Print(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

}

func Println(a ...interface{}) {

	fmt.Println(a...)

}

func Printf(format string, a ...interface{}) {

	fmt.Printf(format, a...)

}

func init() {

}
