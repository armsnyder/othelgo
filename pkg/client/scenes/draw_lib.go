package scenes

import (
	"fmt"
	"math"
	"strings"

	"github.com/nsf/termbox-go"
)

var gameBoyWidth, gameBoyHeight = 96, 24

// draw is a general function for drawing to the terminal.
// The text can be a rune, a string, or a multiline string, which will be drawn relative to the
// specified anchor. This function should always be used instead of termbox.SetCell().
func draw(anchor anchor, color color, text interface{}) {
	switch t := text.(type) {
	case rune:
		positionX, positionY, _, _ := anchor()
		fg, bg := color()
		termbox.SetCell(positionX, positionY, t, fg, bg)

	case string:
		rows := strings.Split(t, "\n")
		textHeight := len(rows)
		textWidth := maxLength(rows)
		_, _, drawDirectionX, drawDirectionY := anchor()
		offsetX := int(math.Round(float64(textWidth) * drawDirectionX))
		offsetY := int(math.Round(float64(textHeight) * drawDirectionY))

		for i, row := range rows {
			for j, ch := range []rune(row) { // Converting to []rune first gets us tight alignment.
				draw(offset(anchor, offsetX+j, offsetY+i), color, ch)
			}
		}

	default:
		panic(fmt.Errorf("unsupported draw text type %T", text))
	}
}

// setCursor uses an anchor to determine the position of the cursor, so it can be used in
// conjunction with draw to place the cursor.
func setCursor(anchor anchor) {
	x, y, _, _ := anchor()
	termbox.SetCursor(x, y-1)
}

// anchor defines a position offset and direction that can be used for drawing.
type anchor func() (positionX, positionY int, drawDirectionX, drawDirectionY float64)

// origin is an anchor on the top-left corner of the terminal window.
func origin() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	return 0, 0, 0, 0
}

// topLeft is an anchor on the top-left corner of the game window.
func topLeft() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	leftX := (termWidth - gameBoyWidth) / 2
	topY := (termHeight - gameBoyHeight) / 2
	// Add an inner margin while also ensuring the text is always on-screen.
	return max(0, leftX+5), max(0, topY+2), 0, 0
}

// topRight is an anchor on the top-right corner of the game window.
func topRight() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	rightX := (termWidth + gameBoyWidth) / 2
	topY := (termHeight - gameBoyHeight) / 2
	// Add an inner margin while also ensuring the text is always on-screen.
	return min(termWidth, rightX-3), max(0, topY+2), -1, 0
}

// botRight is an anchor on the bottom-right corner of the game window.
func botRight() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	rightX := (termWidth + gameBoyWidth) / 2
	botY := (termHeight + gameBoyHeight) / 2
	// Add an inner margin while also ensuring the text is always on-screen.
	return min(termWidth, rightX-3), min(termHeight, botY-1) - 1, -1, 0
}

// center is an anchor the center-middle of the terminal that draws from the center outward.
func center() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	centerX := termWidth / 2
	centerY := termHeight / 2
	return centerX, centerY, -0.5, -0.5
}

// centerLeft is an anchor the center-middle of the terminal that draws toward the left side.
func centerLeft() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	positionX, positionY, _, _ = center()
	return positionX, positionY, -1, -0.5
}

// centerRight is an anchor the center-middle of the terminal that draws toward the right side.
func centerRight() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	positionX, positionY, _, _ = center()
	return positionX, positionY, 0, -0.5
}

// centerTop is an anchor the center-middle of the terminal that draws toward the top.
func centerTop() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	positionX, positionY, _, _ = center()
	return positionX, positionY, -0.5, -1
}

// offset returns an anchor that is offset from the specified anchor by a specified position.
func offset(anchor anchor, x, y int) anchor {
	return func() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
		positionX, positionY, drawDirectionX, drawDirectionY = anchor()
		positionX += x
		positionY += y
		return
	}
}

// color defines foreground and background attributes for drawing.
type color func() (fg, bg termbox.Attribute)

// normal is the default color.
func normal() (fg, bg termbox.Attribute) {
	return termbox.ColorDefault, termbox.ColorDefault
}

// inverted is a color that inverts the text and background color to appear as a highlight.
func inverted() (fg, bg termbox.Attribute) {
	return termbox.ColorBlack, termbox.ColorWhite
}

// magenta is a magenta color.
func magenta() (fg, bg termbox.Attribute) {
	return termbox.ColorMagenta, termbox.ColorDefault
}

// green is not a creative color.
func green() (fg, bg termbox.Attribute) {
	return termbox.ColorGreen, termbox.ColorDefault
}

func drawGameBoyBorder() {
	termWidth, termHeight := termbox.Size()

	topY := (termHeight - gameBoyHeight) / 2
	bottomY := (termHeight + gameBoyHeight) / 2
	leftX := (termWidth - gameBoyWidth) / 2
	rightX := (termWidth + gameBoyWidth) / 2

	borderRunes := []rune{'🎃', '🧟', '🔮', '🧛', '🍬', '👻'}

	for i := 0; i < gameBoyWidth/2; i++ {
		ch := borderRunes[i%len(borderRunes)]
		draw(offset(origin, leftX+i*2, topY), normal, ch)
		draw(offset(origin, leftX+i*2, bottomY), normal, ch)
	}

	for i := 0; i <= gameBoyHeight; i++ {
		ch := borderRunes[i%len(borderRunes)]
		draw(offset(origin, leftX, topY+i), normal, ch)
		draw(offset(origin, rightX, topY+i), normal, ch)
	}
}

func drawSplash() {
	draw(offset(centerTop, 0, 1), normal, `
       _   _          _             
  ___ | |_| |__   ___| | __ _  ___  
 / _ \| __| '_ \ / _ \ |/ _`+"`"+` |/ _ \ 
| (_) | |_| | | |  __/ | (_| | (_) |
 \___/ \__|_| |_|\___|_|\__, |\___/ 
                        |___/       
`)
}
