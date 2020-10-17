package client

import (
	"fmt"
	"unicode"

	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/messages"
)

type gameData struct {
	curSquareX int
	curSquareY int
	board      messages.Board
}

func gameScene(changeScene func(string), sendMessage func(interface{}), onError func(error)) (onMessage func(messages.AnyMessage), onEvent func(termbox.Event)) {
	var gameData gameData
	curPlayer := 1

	drawGame(gameData, onError)

	onMessage = func(message messages.AnyMessage) {
		switch m := message.Message.(type) {
		case *messages.UpdateBoardMessage:
			gameData.board = m.Board

		default:
			onError(fmt.Errorf("unhandled message type %T", m))
		}

		drawGame(gameData, onError)
	}

	onEvent = func(event termbox.Event) {
		updateSelection(event, &gameData)

		if event.Key == termbox.KeyEnter {
			gameData.board[gameData.curSquareX][gameData.curSquareY] = curPlayer

			message := messages.NewPlaceDiskMessage(curPlayer, gameData.curSquareX, gameData.curSquareY)
			sendMessage(message)

			curPlayer = curPlayer%2 + 1
		}

		drawGame(gameData, onError)
	}

	return onMessage, onEvent
}

func updateSelection(event termbox.Event, drawData *gameData) {
	dx, dy := 0, 0

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

	drawData.curSquareX = clamp(drawData.curSquareX+dx, 0, messages.BoardSize)
	drawData.curSquareY = clamp(drawData.curSquareY+dy, 0, messages.BoardSize)
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

func drawGame(drawData gameData, onError func(error)) {
	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		onError(err)
	}

	squareHeight := 2
	squareWidth := 5

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
			termbox.SetCell(x, y, value, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	// Pieces
	for x := 0; x < messages.BoardSize; x++ {
		for y := 0; y < messages.BoardSize; y++ {
			cellX := squareWidth/2 + squareWidth*x
			cellY := squareHeight/2 + squareHeight*y

			var color termbox.Attribute

			switch drawData.board[x][y] {
			case 1:
				color = termbox.ColorMagenta
			case 2:
				color = termbox.ColorGreen
			default:
				continue
			}

			termbox.SetCell(cellX, cellY, '⬤', color, termbox.ColorDefault)
			termbox.SetCell(cellX+1, cellY, ' ', color, termbox.ColorDefault) // Prevent half-circle on some terminals.
		}
	}

	termbox.SetCursor(
		squareWidth/2+squareWidth*drawData.curSquareX,
		squareHeight/2+squareHeight*drawData.curSquareY,
	)

	if err := termbox.Flush(); err != nil {
		onError(err)
	}
}
