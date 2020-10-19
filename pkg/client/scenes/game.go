package scenes

import (
	"fmt"

	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/common"
)

type Game struct {
	scene
	player     int
	curSquareX int
	curSquareY int
	board      common.Board
	p1Score    int
	p2Score    int
}

func (g *Game) Setup(changeScene ChangeScene, sendMessage SendMessage, setupContext SceneContext) error {
	if err := g.scene.Setup(changeScene, sendMessage, setupContext); err != nil {
		return err
	}

	g.player = setupContext["player"].(int)

	var message interface{}
	if g.player == 1 {
		message = common.NewNewGameMessage()
	} else {
		message = common.NewJoinGameMessage()
	}

	return sendMessage(message)
}

func (g *Game) OnMessage(message common.AnyMessage) error {
	if m, ok := message.Message.(*common.UpdateBoardMessage); ok {
		g.board = m.Board
		g.p1Score, g.p2Score = common.KeepScore(g.board)
	}

	return nil
}

func (g *Game) OnTerminalEvent(event termbox.Event) error {
	dx, dy := getDirectionPressed(event)
	g.curSquareX = clamp(g.curSquareX+dx, 0, common.BoardSize)
	g.curSquareY = clamp(g.curSquareY+dy, 0, common.BoardSize)

	if event.Key == termbox.KeyEnter {
		board, updated := common.ApplyMove(g.board, g.curSquareX, g.curSquareY, g.player)
		if updated {
			g.board = board
			message := common.NewPlaceDiskMessage(g.player, g.curSquareX, g.curSquareY)
			if err := g.SendMessage(message); err != nil {
				return err
			}
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
	offset := 21
	drawString(youAreText, 0, 0, termbox.ColorDefault, termbox.ColorDefault)
	drawDisk(g.player, len(youAreText), 0)
	drawString(fmt.Sprintf("Score:    %2d      %2d", g.p1Score, g.p2Score), offset, 0, termbox.ColorDefault, termbox.ColorDefault)
	drawDisk(1, offset+7, 0)
	drawDisk(2, offset+15, 0)

	var (
		squareHeight = 2
		squareWidth  = 5
		yOffset      = 2
	)

	// Outline
	for x := 0; x < common.BoardSize*squareWidth+1; x++ {
		for y := 0; y < common.BoardSize*squareHeight+1; y++ {
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
	for i := 0; i < common.BoardSize; i++ {
		for j := 0; j < common.BoardSize; j++ {
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
