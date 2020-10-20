package scenes

import (
	"unicode"

	"github.com/nsf/termbox-go"
)

func getDirectionPressed(event termbox.Event) (dx, dy int) {
	switch event.Key {
	case termbox.KeyArrowLeft:
		dx = -1
	case termbox.KeyArrowRight:
		dx = 1
	case termbox.KeyArrowUp:
		dy = -1
	case termbox.KeyArrowDown:
		dy = 1
	}

	switch unicode.ToLower(event.Ch) {
	case 'a':
		dx = -1
	case 'd':
		dx = 1
	case 'w':
		dy = -1
	case 's':
		dy = 1
	}

	return dx, dy
}

func maxLength(ss []string) int {
	result := 0
	for _, s := range ss {
		result = max(result, len(s))
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(val, min, max int) int {
	switch {
	case val < min:
		return min
	case val >= max:
		return max - 1
	default:
		return val
	}
}
