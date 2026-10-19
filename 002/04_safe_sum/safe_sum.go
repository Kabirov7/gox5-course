package safe_sum

import (
	"math"
)

// SafeSum последовательно суммирует транзакции, пропуская те, добавление
// которых привело бы к переполнению int64. skipped — индексы пропущенных
// элементов исходного среза, в порядке появления.
func SafeSum(transactions []int64) (total int64, skipped []int) {
	for i, v := range transactions {
		if v > 0 && total > math.MaxInt64-v {
			skipped = append(skipped, i)
			continue
		}

		if v < 0 && total < math.MinInt64-v {
			skipped = append(skipped, i)
			continue
		}
		total += v
	}
	return total, skipped
}
