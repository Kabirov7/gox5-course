package main

import (
	"fmt"
)

func FizzBuzz() {
	for i := 1; i <= 30; i++ {

		if i%3 == 0 && i%5 != 0 {
			fmt.Println("Fizz")
			continue
		}

		if i%3 != 0 && i%5 == 0 {
			fmt.Println("Buzz")
			continue
		}

		if i%3 == 0 && i%5 == 0 {
			fmt.Println("FizzBuzz")
		} else {
			fmt.Println(i)
		}
	}
}

func DayStat(orders []int64) (min, max, total, avg int64) {
	min = orders[0]
	max = orders[0]
	for i := 0; i < len(orders); i++ {
		if min > orders[i] {
			min = orders[i]
		}
		if max < orders[i] {
			max = orders[i]
		}
		total += orders[i]
	}
	avg = total / int64(len(orders))
	return min, max, total, avg
}

func getOrderCategory(amount int) {
	// Сначала проверяем на некорректный ввод (отрицательная сумма)
	if amount < 0 {
		fmt.Println("Ошибка: сумма заказа не может быть отрицательной")
		return // Прерываем выполнение функции
	}

	var category string

	// Определяем категорию. Проверки идут строго сверху вниз.
	switch {
	case amount < 10000:
		// Сюда попадут суммы от 0 до 9999
		category = "small"
	case amount < 50000:
		// Сюда попадут суммы от 10000 до 49999
		category = "regular"
	default:
		// Сюда попадут все суммы от 50000 и выше
		category = "large"
	}

	// Выводим результат
	fmt.Printf("Сумма заказа: %d коп.\nКатегория: %s\n", amount, category)
}

func main() {
	FizzBuzz()
	orders := []int64{
		12000,
		7500,
		31000,
		4500,
	}
	min, max, total, avg := DayStat(orders)
	fmt.Println(min, max, total, avg)
	testAmounts := []int{
		7500,  // обычный small
		25000, // обычный regular
		75000, // обычный large
		0,     // граница small
		9999,  // граница small (последнее значение)
		10000, // граница regular (первое значение)
		49999, // граница regular (последнее значение)
		50000, // граница large (первое значение)
		-1,    // некорректная сумма
		-500,  // некорректная сумма
	}

	// Запускаем проверку для каждого значения
	for _, amount := range testAmounts {
		fmt.Println("-----------------------")
		getOrderCategory(amount)
	}
}
