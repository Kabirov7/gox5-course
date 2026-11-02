package analytics

type Sale struct {
	Product          string
	Category         string
	Quantity         int
	UnitPriceKopecks int64
	Cancelled        bool
}

type Report struct {
	RevenueKopecks    int64
	UnitsSold         int
	RevenueByCategory map[string]int64
	TopProduct        string
	InvalidSales      int
}

func AnalyzeSales(sales []Sale) Report {
	// TODO: реализуйте расчёт отчёта по правилам задания.
	res := Report{
		RevenueKopecks:    0,
		UnitsSold:         0,
		RevenueByCategory: make(map[string]int64),
		TopProduct:        "",
		InvalidSales:      0,
	}
	if len(sales) == 0 {
		return res
	}
	topProducts := make(map[string]int)
	for _, s := range sales {
		if s.Cancelled {
			continue
		}
		if len(s.Product) == 0 || len(s.Category) == 0 || s.Quantity <= 0 || s.UnitPriceKopecks <= 0 {
			res.InvalidSales++
			continue
		}
		res.RevenueKopecks += s.UnitPriceKopecks * int64(s.Quantity)
		res.UnitsSold += s.Quantity
		if _, ok := res.RevenueByCategory[s.Category]; !ok {
			res.RevenueByCategory[s.Category] = 0
		}
		res.RevenueByCategory[s.Category] += s.UnitPriceKopecks * int64(s.Quantity)
		if _, ok := topProducts[s.Product]; !ok {
			topProducts[s.Product] = 0
		}
		topProducts[s.Product] += s.Quantity
	}

	maxCount := 0

	for k, v := range topProducts {
		if v > maxCount {
			res.TopProduct = k
			maxCount = v
		}
	}

	return res
}
