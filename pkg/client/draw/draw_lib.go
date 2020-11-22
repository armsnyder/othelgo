package draw

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

// Bordered game window area size.
const gameBoyWidth, gameBoyHeight = 96, 24

// Margins inside the bordered game window.
const marginX, marginY = 3, 1

// Draw is a general function for drawing to the terminal.
// The text can be a rune, a string, or a multiline string, which will be drawn relative to the
// specified anchor. This function should always be used instead of termbox.SetCell().
func Draw(anchor Anchor, color Color, text interface{}) {
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
				Draw(Offset(anchor, offsetX+j, offsetY+i), color, ch)
			}
		}

	default:
		panic(fmt.Errorf("unsupported Draw text type %T", text))
	}
}

// SetCursor uses an Anchor to determine the position of the cursor, so it can be used in
// conjunction with Draw to place the cursor.
func SetCursor(anchor Anchor) {
	x, y, _, _ := anchor()
	termbox.SetCursor(x, y-1)
}

// Anchor defines a position offset and direction that can be used for drawing.
type Anchor func() (positionX, positionY int, drawDirectionX, drawDirectionY float64)

// Origin is an Anchor on the top-left corner of the terminal window.
func Origin() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	return 0, 0, 0, 0
}

// TopLeft is an Anchor on the top-left corner of the game window.
func TopLeft() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	leftX := (termWidth - gameBoyWidth) / 2
	topY := (termHeight - gameBoyHeight) / 2
	// Add an inner margin while also ensuring the text is always on-screen.
	return max(0, leftX+marginX+2), max(0, topY+marginY+1), 0, 0
}

// TopRight is an Anchor on the top-right corner of the game window.
func TopRight() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	rightX := (termWidth + gameBoyWidth) / 2
	topY := (termHeight - gameBoyHeight) / 2
	// Add an inner margin while also ensuring the text is always on-screen.
	return min(termWidth, rightX-marginX), max(0, topY+marginY+1), -1, 0
}

// BotRight is an Anchor on the bottom-right corner of the game window.
func BotRight() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	rightX := (termWidth + gameBoyWidth) / 2
	botY := (termHeight + gameBoyHeight) / 2
	// Add an inner margin while also ensuring the text is always on-screen.
	return min(termWidth, rightX-marginX), min(termHeight, botY-marginY) - 1, -1, 0
}

// Center is an Anchor in the center-middle of the terminal that draws from the center outward.
func Center() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	centerX := termWidth / 2
	centerY := termHeight / 2
	return centerX, centerY, -0.5, -0.5
}

// CenterLeft is an Anchor in the center-middle of the terminal that draws toward the left side.
func CenterLeft() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	positionX, positionY, _, _ = Center()
	return positionX, positionY, -1, -0.5
}

// CenterRight is an Anchor in the center-middle of the terminal that draws toward the right side.
func CenterRight() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	positionX, positionY, _, _ = Center()
	return positionX, positionY, 0, -0.5
}

// CenterTop is an Anchor in the center-middle of the terminal that draws toward the top.
func CenterTop() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	positionX, positionY, _, _ = Center()
	return positionX, positionY, -0.5, -1
}

// MiddleRight is an Anchor that is vertically centered and on the right edge of the game window.
func MiddleRight() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
	termWidth, termHeight := termbox.Size()
	rightX := (termWidth + gameBoyWidth) / 2
	centerY := termHeight / 2
	// Add an inner margin while also ensuring the text is always on-screen.
	return min(termWidth, rightX-marginX), centerY, -1, 0
}

// Offset returns an Anchor that is cffset from the specified Anchor by a specified position.
func Offset(anchor Anchor, x, y int) Anchor {
	return func() (positionX, positionY int, drawDirectionX, drawDirectionY float64) {
		positionX, positionY, drawDirectionX, drawDirectionY = anchor()
		positionX += x
		positionY += y
		return
	}
}

// Color defines foreground and background attributes for drawing.
type Color func() (fg, bg termbox.Attribute)

// Normal is the default Color.
func Normal() (fg, bg termbox.Attribute) {
	return termbox.ColorDefault, termbox.ColorDefault
}

// Inverted is a Color that inverts the text and background color to appear as a highlight.
func Inverted() (fg, bg termbox.Attribute) {
	return termbox.ColorBlack, termbox.ColorWhite
}

// Magenta is a magenta Color.
func Magenta() (fg, bg termbox.Attribute) {
	return termbox.ColorMagenta, termbox.ColorDefault
}

// Green is not a creative Color.
func Green() (fg, bg termbox.Attribute) {
	return termbox.ColorGreen, termbox.ColorDefault
}

func Border(decoration string) {
	if decoration == "" {
		return
	}

	termWidth, termHeight := termbox.Size()

	topY := (termHeight - gameBoyHeight) / 2
	bottomY := (termHeight + gameBoyHeight) / 2
	leftX := (termWidth - gameBoyWidth) / 2
	rightX := (termWidth + gameBoyWidth) / 2

	borderRunes := []rune(decoration)

	for i := 0; i < gameBoyWidth/2; i++ {
		ch := borderRunes[i%len(borderRunes)]
		Draw(Offset(Origin, leftX+i*2, topY), Normal, ch)
		Draw(Offset(Origin, leftX+i*2, bottomY), Normal, ch)
	}

	for i := 0; i <= gameBoyHeight; i++ {
		ch := borderRunes[i%len(borderRunes)]
		Draw(Offset(Origin, leftX, topY+i), Normal, ch)
		Draw(Offset(Origin, rightX, topY+i), Normal, ch)
	}
}

func maxLength(ss []string) int {
	result := 0
	for _, s := range ss {
		result = max(result, utf8.RuneCountInString(s))
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
