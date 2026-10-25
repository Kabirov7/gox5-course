package main

import (
	"fmt"
)

func Increment(value *int) {
	if value == nil {
		return
	}
	*value++
}

func main() {
	number := 10
	Increment(&number)
	fmt.Println(number) // 11

	Increment(&number)
	fmt.Println(number) // 12

	Increment(nil)

	Increment(&number)
	fmt.Println(number) // 13
}
