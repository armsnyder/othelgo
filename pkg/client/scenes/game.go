package scenes

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/armsnyder/othelgo/pkg/common"

	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/client/draw"
	"github.com/armsnyder/othelgo/pkg/messages"
)

type Game struct {
	scene
	player       common.Disk
	curSquareX   int
	curSquareY   int
	board        common.Board
	p1Score      int
	p2Score      int
	confetti     confetti
	nickname     string
	host         string
	opponent     string
	whoseTurn    common.Disk
	multiplayer  bool
	difficulty   int
	alertMessage string
	prevX        int
	prevY        int
}

func (g *Game) Setup(changeScene ChangeScene, sendMessage SendMessage) error {
	if err := g.scene.Setup(changeScene, sendMessage); err != nil {
		return err
	}

	if g.multiplayer && g.player == 1 {
		g.alertMessage = "Waiting for opponent"
	}

	var message interface{}
	if g.multiplayer {
		if g.player == 1 {
			message = messages.HostGame{Nickname: g.nickname}
		} else {
			message = messages.JoinGame{Nickname: g.nickname, Host: g.host}
		}
	} else {
		message = messages.StartSoloGame{Nickname: g.nickname, Difficulty: g.difficulty}
	}

	return sendMessage(message)
}

func (g *Game) OnMessage(message interface{}) error {
	switch m := message.(type) {
	case *messages.UpdateBoard:
		g.board = m.Board
		g.whoseTurn = m.Player
		g.p1Score = m.P1Score
		g.p2Score = m.P2Score
		if m.X >= 0 && m.Y >= 0 {
			g.prevX = m.X
			g.prevY = m.Y
		}
	case *messages.GameOver:
		g.alertMessage = m.Message
	case *messages.Joined:
		g.alertMessage = ""
		if g.nickname == g.host {
			g.opponent = m.Nickname
		}
	}

	return nil
}

func (g *Game) OnTerminalEvent(event termbox.Event) error {
	if unicode.ToUpper(event.Ch) == 'M' {
		g.OnQuit()
		return g.ChangeScene(&Menu{nickname: g.nickname})
	}

	if g.alertMessage != "" {
		return nil
	}

	dx, dy := getDirectionPressed(event)
	g.curSquareX = clamp(g.curSquareX+dx, 0, common.BoardSize)
	g.curSquareY = clamp(g.curSquareY+dy, 0, common.BoardSize)

	if event.Key == termbox.KeyEnter && g.whoseTurn == g.player {
		board, updated := common.ApplyMove(g.board, g.curSquareX, g.curSquareY, g.player)
		if updated {
			g.board = board
			message := messages.PlaceDisk{
				Nickname: g.nickname,
				Host:     g.host,
				X:        g.curSquareX,
				Y:        g.curSquareY,
			}
			if err := g.SendMessage(message); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Game) OnQuit() {
	if err := g.SendMessage(messages.LeaveGame{Nickname: g.nickname, Host: g.host}); err != nil {
		log.Print(err)
	}
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
	g.drawScore()
	draw.Draw(draw.TopRight, draw.Normal, fmt.Sprintf("Your name is %s!", strings.ToUpper(g.nickname)))
	draw.Draw(draw.BotRight, draw.Normal, "[M] MENU  [Q] QUIT")
	drawBoardOutline()
	g.drawDisks()
	g.drawCursor()
	g.confetti.draw()
	g.drawAlert()
	if g.player == g.whoseTurn && (g.p1Score+g.p2Score > 4) {
		g.highlightMove(g.prevX, g.prevY)
	}
}

var playerColors = map[common.Disk]draw.Color{1: draw.Magenta, 2: draw.Green}

func drawDisk(anchor draw.Anchor, player common.Disk) {
	// The extra space prevents a half-circle on some terminals.
	draw.Draw(anchor, playerColors[player], "⬤ ")
}

func (g *Game) highlightMove(x, y int) {
	draw.Draw(draw.Offset(draw.Center, ((x+1-common.BoardSize/2)*squareWidth)-4, (y+1-common.BoardSize/2)*squareHeight), draw.Normal, "[")
	draw.Draw(draw.Offset(draw.Center, ((x+1-common.BoardSize/2)*squareWidth)-1, (y+1-common.BoardSize/2)*squareHeight), draw.Normal, "]")
}

var (
	squareWidth  = 5
	squareHeight = 2
)

func (g *Game) drawScore() {
	var p1Name, p2Name string
	if g.player == 1 {
		p1Name = strings.ToUpper(g.nickname)
		p2Name = strings.ToUpper(g.opponent)
	} else {
		p1Name = strings.ToUpper(g.opponent)
		p2Name = strings.ToUpper(g.nickname)
	}

	// P1 Name and Score
	drawDisk(draw.Offset(draw.MiddleLeft, 4, -1), 1)
	draw.Draw(draw.Offset(draw.MiddleLeft, 7, -1), draw.Normal, fmt.Sprintf("%s: %-2d", p1Name, g.p1Score))

	// P2 Name and Score
	drawDisk(draw.Offset(draw.MiddleLeft, 4, 1), 2)
	draw.Draw(draw.Offset(draw.MiddleLeft, 7, 1), draw.Normal, fmt.Sprintf("%s: %-2d", p2Name, g.p2Score))

	// Current turn indicator
	if !common.GameOver(g.board) {
		var yOffset int
		if g.whoseTurn == 1 {
			yOffset = 0
		} else {
			yOffset = 2
		}
		draw.Draw(draw.Offset(draw.MiddleLeft, 4, yOffset), draw.Normal, "﹌")
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
			// Crossing.
			case y%squareHeight == 0 && x%squareWidth == 0:
				switch {
				// Top row crossing.
				case y == -boardHeight/2:
					switch x {
					case -boardWidth / 2:
						value = '┌'
					case boardWidth / 2:
						value = '┐'
					default:
						value = '┬'
					}

				// Bottom row crossing.
				case y == boardHeight/2:
					switch x {
					case -boardWidth / 2:
						value = '└'
					case boardWidth / 2:
						value = '┘'
					default:
						value = '┴'
					}

				// Left side crossing.
				case x == -boardWidth/2:
					value = '├'

				// Right side crossing.
				case x == boardWidth/2:
					value = '┤'

				// Inner crossing.
				default:
					value = '┼'
				}

			case y%squareHeight == 0:
				value = '─'

			case x%squareWidth == 0:
				value = '│'
			}

			draw.Draw(draw.Offset(draw.Center, x, y), draw.Normal, value)
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

			x := (i+1-common.BoardSize/2)*squareWidth - 2
			y := (j + 1 - common.BoardSize/2) * squareHeight

			drawDisk(draw.Offset(draw.Center, x, y), player)
		}
	}
}

func (g *Game) drawCursor() {
	if common.GameOver(g.board) || g.whoseTurn != g.player || g.alertMessage != "" {
		termbox.HideCursor()
	} else {
		x := (g.curSquareX+1-common.BoardSize/2)*squareWidth - 3
		y := (g.curSquareY + 1 - common.BoardSize/2) * squareHeight

		draw.SetCursor(draw.Offset(draw.Center, x, y))
	}
}

func (g *Game) drawAlert() {
	if g.alertMessage == "" {
		return
	}

	var sb strings.Builder

	writeLine := func(first rune, content string, last rune) {
		sb.WriteRune(first)
		sb.WriteString(content)
		sb.WriteRune(last)
		sb.WriteRune('\n')
	}

	fillLine := func(first, ch, last rune) {
		content := make([]rune, len(g.alertMessage)+4)
		for i := 0; i < len(content); i++ {
			content[i] = ch
		}
		writeLine(first, string(content), last)
	}

	fillLine('╔', '═', '╗')
	fillLine('║', ' ', '║')
	writeLine('║', fmt.Sprintf("  %s  ", g.alertMessage), '║')
	fillLine('║', ' ', '║')
	fillLine('╚', '═', '╝')

	draw.Draw(draw.Center, draw.Normal, sb.String())
}
