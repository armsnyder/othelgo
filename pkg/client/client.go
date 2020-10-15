package client

import (
	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/common"
)

type drawData struct {
	curSquareX int
	curSquareY int
	board      common.Board
}

func Run() error {
	c, _, err := websocket.DefaultDialer.Dial("wss://1y9vcb5geb.execute-api.us-west-2.amazonaws.com/development", nil)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	terminalEvents := make(chan termbox.Event)
	go receiveTerminalEvents(terminalEvents)

	var drawData drawData
	curPlayer := 1

	if err := draw(drawData); err != nil {
		return err
	}

	for event := range terminalEvents {
		if shouldInterrupt(event) {
			termbox.Interrupt()
		}

		updateSelection(event, &drawData)

		if event.Key == termbox.KeyEnter {
			drawData.board[drawData.curSquareX][drawData.curSquareY] = curPlayer
			curPlayer %= 2
			curPlayer++
		}

		if err := draw(drawData); err != nil {
			return err
		}
	}

	return nil
}

func receiveTerminalEvents(ch chan<- termbox.Event) {
	for {
		event := termbox.PollEvent()

		switch event.Type {
		case termbox.EventError, termbox.EventInterrupt:
			close(ch)
		default:
			ch <- event
		}
	}
}

func shouldInterrupt(event termbox.Event) bool {
	return event.Ch == 'q' || event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyEsc
}

func updateSelection(event termbox.Event, drawData *drawData) {
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

	switch event.Ch {
	case 'a', 'A':
		dx = -1
	case 'd', 'D':
		dx = 1
	case 'w', 'W':
		dy = -1
	case 's', 'S':
		dy = 1
	}

	drawData.curSquareX = clamp(drawData.curSquareX+dx, 0, common.BoardSize)
	drawData.curSquareY = clamp(drawData.curSquareY+dy, 0, common.BoardSize)
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

func draw(drawData drawData) error {
	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		return err
	}

	squareHeight := 2
	squareWidth := 5

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
			termbox.SetCell(x, y, value, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	// Pieces
	for x := 0; x < common.BoardSize; x++ {
		for y := 0; y < common.BoardSize; y++ {
			cellX := squareWidth/2 + squareWidth*x
			cellY := squareHeight/2 + squareHeight*y

			switch drawData.board[x][y] {
			case 1:
				termbox.SetCell(cellX, cellY, '⬤', termbox.ColorMagenta, termbox.ColorDefault)
			case 2:
				termbox.SetCell(cellX, cellY, '⬤', termbox.ColorGreen, termbox.ColorDefault)
			}
		}
	}

	termbox.SetCursor(
		squareWidth/2+squareWidth*drawData.curSquareX,
		squareHeight/2+squareHeight*drawData.curSquareY,
	)

	return termbox.Flush()
}
