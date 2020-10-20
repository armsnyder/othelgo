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
	confetti   confetti
}

func (g *Game) Setup(changeScene ChangeScene, sendMessage SendMessage) error {
	if err := g.scene.Setup(changeScene, sendMessage); err != nil {
		return err
	}

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
		updated := common.ApplyMove(&g.board, g.curSquareX, g.curSquareY, g.player)
		if updated {
			message := common.NewPlaceDiskMessage(g.player, g.curSquareX, g.curSquareY)
			if err := g.SendMessage(message); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Game) Tick() bool {
	if !common.GameOver(g.board) {
		return false
	}

	p1, p2 := common.KeepScore(g.board)
	switch {
	case g.player == 1 && p2 > p1:
		return false
	case g.player == 2 && p1 > p2:
		return false
	}

	g.confetti.tick()
	return true
}

func (g *Game) Draw() {
	drawGameBoyBorder()
	g.drawYouAre()
	g.drawScore()
	drawBoardOutline()
	g.drawDisks()
	g.drawCursor()
	g.confetti.draw()
}

var playerColors = map[int]termbox.Attribute{
	1: termbox.ColorMagenta,
	2: termbox.ColorGreen,
}

func drawDisk(player, x, y int) {
	color := playerColors[player]
	termbox.SetCell(x, y, '⬤', color, termbox.ColorDefault)
	termbox.SetCell(x+1, y, ' ', color, termbox.ColorDefault) // Prevent half-circle on some terminals.
}

func (g *Game) drawYouAre() {
	topY, _, leftX, _ := corners()

	youAreText := "You are: "
	drawStringDefault(youAreText, leftX+5, topY+2)
	drawDisk(g.player, leftX+5+len(youAreText), topY+2)
}

var (
	squareWidth  = 5
	squareHeight = 2
)

func (g *Game) drawScore() {
	topY, _, _, rightX := corners()

	var (
		p2ScoreXOffset = rightX - 5
		p2DiskXOffset  = p2ScoreXOffset - 3
		p1ScoreXOffset = p2DiskXOffset - 4
		p1DiskXOffset  = p1ScoreXOffset - 3
		scoreText      = "Score: "
		scoreXOffset   = p1DiskXOffset - len(scoreText)
	)

	drawStringDefault(scoreText, scoreXOffset, topY+2)
	drawStringDefault(fmt.Sprintf("%2d", g.p1Score), p1ScoreXOffset, topY+2)
	drawStringDefault(fmt.Sprintf("%2d", g.p2Score), p2ScoreXOffset, topY+2)
	drawDisk(1, p1DiskXOffset, topY+2)
	drawDisk(2, p2DiskXOffset, topY+2)

	// Current turn indicator
	if !common.GameOver(g.board) {
		if common.WhoseTurn(g.board) == 1 {
			drawStringDefault("﹌", p1DiskXOffset, topY+3)
		} else {
			drawStringDefault("﹌", p2DiskXOffset, topY+3)
		}
	}
}

func drawBoardOutline() {
	termWidth, termHeight := termbox.Size()

	var (
		boardWidth  = common.BoardSize * squareWidth
		boardHeight = common.BoardSize * squareHeight
	)

	offsetX, offsetY := (termWidth-boardWidth)/2, (termHeight-boardHeight)/2

	// Outline
	for x := 0; x <= boardWidth; x++ {
		for y := 0; y <= boardHeight; y++ {
			var value rune

			switch {
			case y%squareHeight == 0 && x%squareWidth == 0:
				value = '+'
			case y%squareHeight == 0:
				value = '-'
			case x%squareWidth == 0:
				value = '|'
			}

			termbox.SetCell(offsetX+x, offsetY+y, value, termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (g *Game) drawDisks() {
	termWidth, termHeight := termbox.Size()

	var (
		boardWidth  = common.BoardSize * squareWidth
		boardHeight = common.BoardSize * squareHeight
	)

	offsetX, offsetY := (termWidth-boardWidth)/2, (termHeight-boardHeight)/2

	for i := 0; i < common.BoardSize; i++ {
		for j := 0; j < common.BoardSize; j++ {
			player := g.board[i][j]
			if player == 0 {
				continue
			}

			x := offsetX + squareWidth/2 + squareWidth*i
			y := offsetY + squareHeight/2 + squareHeight*j

			drawDisk(player, x, y)
		}
	}
}

func (g *Game) drawCursor() {
	termWidth, termHeight := termbox.Size()

	var (
		boardWidth  = common.BoardSize * squareWidth
		boardHeight = common.BoardSize * squareHeight
	)

	offsetX, offsetY := (termWidth-boardWidth)/2, (termHeight-boardHeight)/2

	if common.GameOver(g.board) || common.WhoseTurn(g.board) != g.player {
		termbox.HideCursor()
	} else {
		termbox.SetCursor(
			offsetX+squareWidth/2+squareWidth*g.curSquareX,
			offsetY+squareHeight/2+squareHeight*g.curSquareY,
		)
	}
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
