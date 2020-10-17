package client

import (
	"fmt"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/messages"
)

type drawData struct {
	curSquareX int
	curSquareY int
	board      messages.Board
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

	messageQueue := make(chan messages.AnyMessage)
	messageErrors := make(chan error)
	go receiveMessages(c, messageQueue, messageErrors)

	var drawData drawData
	curPlayer := 1

	if err := draw(drawData); err != nil {
		return err
	}

	for {
		select {
		case event := <-terminalEvents:
			if shouldInterrupt(event) {
				termbox.Interrupt()
				return nil
			}

			updateSelection(event, &drawData)

			if event.Key == termbox.KeyEnter {
				drawData.board[drawData.curSquareX][drawData.curSquareY] = curPlayer

				message := messages.NewPlaceDiskMessage(curPlayer, drawData.curSquareX, drawData.curSquareY)
				if err := c.WriteJSON(message); err != nil {
					return err
				}

				curPlayer = curPlayer%2 + 1
			}

		case anyMessage := <-messageQueue:
			switch m := anyMessage.Message.(type) {
			case *messages.UpdateBoardMessage:
				drawData.board = m.Board

			default:
				return fmt.Errorf("unhandled message type %T", m)
			}

		case err := <-messageErrors:
			return err
		}

		if err := draw(drawData); err != nil {
			return err
		}
	}
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

func receiveMessages(c *websocket.Conn, messageQueue chan<- messages.AnyMessage, messageErrors chan<- error) {
	for {
		var message messages.AnyMessage
		if err := c.ReadJSON(&message); err != nil {
			messageErrors <- err
		}
		messageQueue <- message
	}
}

func shouldInterrupt(event termbox.Event) bool {
	return unicode.ToLower(event.Ch) == 'q' || event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyEsc
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

func draw(drawData drawData) error {
	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		return err
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

	return termbox.Flush()
}
