package scenes

import (
	"strings"
	"unicode"

	"github.com/nsf/termbox-go"
)

var splashText = `
       _   _          _             
  ___ | |_| |__   ___| | __ _  ___  
 / _ \| __| '_ \ / _ \ |/ _` + "`" + ` |/ _ \ 
| (_) | |_| | | |  __/ | (_| | (_) |
 \___/ \__|_| |_|\___|_|\__, |\___/ 
                        |___/       
`

var gameBoyWidth, gameBoyHeight = 96, 24

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

func drawStringDefault(s string, x, y int) {
	drawString(s, x, y, termbox.ColorDefault, termbox.ColorDefault)
}

func drawFromCenter(s string, dx, dy int, fg, bg termbox.Attribute) {
	termWidth, termHeight := termbox.Size()

	rows := strings.Split(s, "\n")
	textWidth := 0
	for _, r := range rows {
		textWidth = max(textWidth, len(r))
	}

	x := termWidth/2 - textWidth/2 + dx
	y := termHeight/2 + dy

	for i, r := range rows {
		drawString(r, x, y+i, fg, bg)
	}
}

func drawUpperRight(s string) {
	topY, _, _, rightX := corners()

	drawStringDefault(s, rightX-len(s)-3, topY+2)
}

func drawGameBoyBorder() {
	topY, bottomY, leftX, rightX := corners()

	borderRunes := []rune{'🎃', '🧟', '🔮', '🧛', '🍬', '👻'}

	for i := 0; i < gameBoyWidth/2; i++ {
		ch := borderRunes[i%len(borderRunes)]
		termbox.SetCell(leftX+i*2, topY, ch, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(leftX+i*2, bottomY, ch, termbox.ColorDefault, termbox.ColorDefault)
	}

	for i := 0; i <= gameBoyHeight; i++ {
		ch := borderRunes[i%len(borderRunes)]
		termbox.SetCell(leftX, topY+i, ch, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(rightX, topY+i, ch, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func corners() (topY, bottomY, leftX, rightX int) {
	termWidth, termHeight := termbox.Size()

	topY = (termHeight - gameBoyHeight) / 2
	bottomY = (termHeight + gameBoyHeight) / 2
	leftX = (termWidth - gameBoyWidth) / 2
	rightX = (termWidth + gameBoyWidth) / 2

	return
}
