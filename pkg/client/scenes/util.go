package scenes

import (
	"unicode"

	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/client/draw"
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

func min(a, b int) int {
	if a < b {
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

func drawSplash() {
	draw.Draw(draw.Offset(draw.CenterTop, 0, 1), draw.Normal, `
       _   _          _             
  ___ | |_| |__   ___| | __ _  ___  
 / _ \| __| '_ \ / _ \ |/ _`+"`"+` |/ _ \ 
| (_) | |_| | | |  __/ | (_| | (_) |
 \___/ \__|_| |_|\___|_|\__, |\___/ 
                        |___/       
`)
}
