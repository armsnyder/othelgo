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

func drawString(s string, x, y int, fg, bg termbox.Attribute) {
	for i, ch := range s {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}

	// Reset terminal colors.
	termbox.SetCell(x+len(s), y, ' ', termbox.ColorDefault, termbox.ColorDefault)
}
