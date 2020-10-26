package scenes

import (
	"fmt"
	"unicode"

	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/common"
)

type Game struct {
	scene
	player     common.Disk
	curSquareX int
	curSquareY int
	board      common.Board
	p1Score    int
	p2Score    int
	confetti   confetti
	nickname   string
	whoseTurn  common.Disk
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
		g.whoseTurn = m.Player
		g.p1Score, g.p2Score = common.KeepScore(g.board)
	}

	return nil
}

func (g *Game) OnTerminalEvent(event termbox.Event) error {
	dx, dy := getDirectionPressed(event)
	g.curSquareX = clamp(g.curSquareX+dx, 0, common.BoardSize)
	g.curSquareY = clamp(g.curSquareY+dy, 0, common.BoardSize)

	if event.Key == termbox.KeyEnter && g.whoseTurn == g.player {
		board, updated := common.ApplyMove(g.board, g.curSquareX, g.curSquareY, g.player)
		if updated {
			g.board = board
			message := common.NewPlaceDiskMessage(g.player, g.curSquareX, g.curSquareY)
			if err := g.SendMessage(message); err != nil {
				return err
			}
		}
	}

	if unicode.ToUpper(event.Ch) == 'M' {
		return g.ChangeScene(&Menu{nickname: g.nickname})
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
	draw(topRight, normal, fmt.Sprintf("Your name is %s!", g.nickname))
	draw(botRight, normal, "[M] MENU  [Q] QUIT")
	drawBoardOutline()
	g.drawDisks()
	g.drawCursor()
	g.confetti.draw()
}

var playerColors = map[common.Disk]color{1: magenta, 2: green}

func drawDisk(anchor anchor, player common.Disk) {
	// The extra space prevents a half-circle on some terminals.
	draw(anchor, playerColors[player], "⬤ ")
}

func (g *Game) drawYouAre() {
	youAreText := "You are: "
	draw(topLeft, normal, youAreText)
	drawDisk(offset(topLeft, len(youAreText), 0), g.player)
}

var (
	squareWidth  = 5
	squareHeight = 2
)

func (g *Game) drawScore() {
	// Text.
	scoreText := "Score: "
	draw(offset(middleRight, len(scoreText)-20, 0), normal, scoreText)

	// P1 score.
	drawDisk(offset(middleRight, -8, 0), 1)
	draw(offset(middleRight, -7, 0), normal, fmt.Sprintf("%2d", g.p1Score))

	// P2 score.
	drawDisk(offset(middleRight, -1, 0), 2)
	draw(middleRight, normal, fmt.Sprintf("%2d", g.p2Score))

	// Current turn indicator
	if !common.GameOver(g.board) {
		var xOffset int
		if g.whoseTurn == 1 {
			xOffset = -9
		} else {
			xOffset = -2
		}
		draw(offset(middleRight, xOffset, 1), normal, "﹌")
	}
}

func drawBoardOutline() {
	var (
		boardWidth  = common.BoardSize * squareWidth
		boardHeight = common.BoardSize * squareHeight
	)

	// Outline
	for x := -boardWidth / 2; x <= boardWidth/2; x++ {
		for y := -boardHeight / 2; y <= boardHeight/2; y++ {
			var value rune

			switch {
			case y%squareHeight == 0 && x%squareWidth == 0:
				value = '+'
			case y%squareHeight == 0:
				value = '-'
			case x%squareWidth == 0:
				value = '|'
			}

			draw(offset(center, x, y), normal, value)
		}
	}
}

func (g *Game) drawDisks() {
	for i := 0; i < common.BoardSize; i++ {
		for j := 0; j < common.BoardSize; j++ {
			player := g.board[i][j]
			if player == 0 {
				continue
			}

			x := (i+1-common.BoardSize/2)*squareWidth - 1
			y := (j + 1 - common.BoardSize/2) * squareHeight

			drawDisk(offset(center, x, y), player)
		}
	}
}

func (g *Game) drawCursor() {
	if common.GameOver(g.board) || g.whoseTurn != g.player {
		termbox.HideCursor()
	} else {
		x := (g.curSquareX+1-common.BoardSize/2)*squareWidth - 3
		y := (g.curSquareY + 1 - common.BoardSize/2) * squareHeight

		setCursor(offset(center, x, y))
	}
}
