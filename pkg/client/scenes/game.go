package scenes

import (
	"fmt"

	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/messages"
)

type Game struct {
	scene
	player     int
	curSquareX int
	curSquareY int
	board      messages.Board
}

func (g *Game) Setup(changeScene ChangeScene, sendMessage SendMessage, setupContext SceneContext) {
	g.scene.Setup(changeScene, sendMessage, setupContext)
	g.player = setupContext["player"].(int)
}

func (g *Game) OnMessage(message messages.AnyMessage) error {
	switch m := message.Message.(type) {
	case *messages.UpdateBoardMessage:
		g.board = m.Board

	default:
		return fmt.Errorf("unhandled message type %T", m)
	}

	return nil
}

func (g *Game) OnTerminalEvent(event termbox.Event) error {
	dx, dy := getDirectionPressed(event)
	g.curSquareX = clamp(g.curSquareX+dx, 0, messages.BoardSize)
	g.curSquareY = clamp(g.curSquareY+dy, 0, messages.BoardSize)

	if event.Key == termbox.KeyEnter {
		g.board[g.curSquareX][g.curSquareY] = g.player

		message := messages.NewPlaceDiskMessage(g.player, g.curSquareX, g.curSquareY)
		if err := g.SendMessage(message); err != nil {
			return err
		}
	}

	return nil
}

func (g *Game) Draw() {
	playerColors := map[int]termbox.Attribute{
		1: termbox.ColorMagenta,
		2: termbox.ColorGreen,
	}

	drawDisk := func(player, x, y int) {
		color := playerColors[player]
		termbox.SetCell(x, y, '⬤', color, termbox.ColorDefault)
		termbox.SetCell(x+1, y, ' ', color, termbox.ColorDefault) // Prevent half-circle on some terminals.
	}

	youAreText := "You are: "
	drawString(youAreText, 0, 0, termbox.ColorDefault, termbox.ColorDefault)
	drawDisk(g.player, len(youAreText), 0)

	var (
		squareHeight = 2
		squareWidth  = 5
		yOffset      = 2
	)

	// Outline
	for x := 0; x < messages.BoardSize*squareWidth+1; x++ {
		for y := 0; y < messages.BoardSize*squareHeight+1; y++ {
			var value rune
			switch {
			case y%squareHeight == 0 && x%squareWidth == 0:
				value = '+'
			case y%squareHeight == 0:
				value = '-'
			case x%squareWidth == 0:
				value = '|'
			}
			termbox.SetCell(x, yOffset+y, value, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	// Pieces
	for i := 0; i < messages.BoardSize; i++ {
		for j := 0; j < messages.BoardSize; j++ {
			player := g.board[i][j]
			if player == 0 {
				continue
			}

			x := squareWidth/2 + squareWidth*i
			y := yOffset + squareHeight/2 + squareHeight*j

			drawDisk(player, x, y)
		}
	}

	termbox.SetCursor(
		squareWidth/2+squareWidth*g.curSquareX,
		yOffset+squareHeight/2+squareHeight*g.curSquareY,
	)
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
