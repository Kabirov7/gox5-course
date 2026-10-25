package main

import (
	"fmt"
)

func StudentScores() map[string]int {
	scores := make(map[string]int, 3)
	scores["anna"] = 3
	scores["boris"] = 0
	scores["vera"] = 5

	scores["anna"] += 2
	delete(scores, "vera")
	return scores
}

func main() {
	scores := StudentScores()
	fmt.Println("Баллы Анны после изменения", scores["anna"])

	boris, ok := scores["boris"]
	fmt.Println("Поиск Бориса: ", boris, ok)

	gleb, ok := scores["gleb"]
	fmt.Println("Поиск Глеба", gleb, ok)

	fmt.Println("Длина после удаления Веры", len(scores))
}
