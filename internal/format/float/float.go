package float

import (
	"strconv"
	"strings"
)

// formatFloat форматирует float64, удаляя лишние нули.
func FormatFloat(value float64) string {
	// Преобразуем число в строку с максимальной точностью.
	str := strconv.FormatFloat(value, 'f', -1, 64)
	// Удаляем лишние нули и точку, если они есть.
	str = strings.TrimRight(str, "0")
	str = strings.TrimRight(str, ".")
	return str
}
