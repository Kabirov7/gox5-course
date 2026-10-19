package unicode_name

import (
	"fmt"
	"unicode"
)

// AnalyzeName считает байты, Unicode-кодовые точки, латинские и кириллические
// буквы и возвращает байтовые позиции начала каждой Unicode-кодовой точки.
// Предполагается, что входная строка содержит корректный UTF-8.
func AnalyzeName(name string) (byteCount, runeCount, latinCount, cyrillicCount int, positions []int) {
	// TODO: используйте range по строке; для классификации букв можно

	for i, v := range name {
		fmt.Println(v)
		fmt.Println(i)
		fmt.Println()
		if unicode.In(v, unicode.Latin) && unicode.IsLetter(v) {
			latinCount++
		}
		if unicode.In(v, unicode.Cyrillic) && unicode.IsLetter(v) {
			cyrillicCount++
		}
		runeCount++
		positions = append(positions, i)
	}
	// использовать стандартный пакет unicode.
	return len(name), runeCount, latinCount, cyrillicCount, positions
}
