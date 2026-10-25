package main

import (
	"fmt"
)

func StudentScores() map[string]int {
	scores := make(map[string]int, 3)
	scores["Анна"] = 3
	scores["Борис"] = 0
	scores["Вера"] = 5

	scores["Анна"] += 2
	delete(scores, "Вера")
	return scores
}

func main() {
	scores := StudentScores()
	fmt.Println("Баллы Анны после изменения", scores["Анна"])

	boris, ok := scores["Борис"]
	fmt.Println("Поиск Бориса: ", boris, ok)

	gleb, ok := scores["Глеб"]
	fmt.Println("Поиск Глеба", gleb, ok)

	fmt.Println("Длина после удаления Веры", len(scores))
}
