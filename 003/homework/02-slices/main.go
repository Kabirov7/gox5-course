package main

import (
	"fmt"
)

func Sum(numbers []int) int {
	if len(numbers) == 0 {
		return 0
	}
	res := 0
	for _, i := range numbers {
		res += i
	}
	return res
}

func main() {
	numbers := []int{
		2,
		4,
		6,
		8,
		10,
	}
	fmt.Println(Sum(numbers)) // 30
	fmt.Println(Sum([]int{})) // 0
	fmt.Println(Sum(nil))     // 0
	fmt.Println(
		Sum(
			[]int{
				-2,
				-1,
				0,
				1,
				2,
			},
		),
	) // 0

	numbers = append(numbers, 12)
	fmt.Println(len(numbers), Sum(numbers))

	shared := numbers
	shared[0] = 100
	fmt.Println(numbers[0], shared[0])
}
