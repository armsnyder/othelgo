package client

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
)

const boardSize = 8

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

	drawBoard()

	if err := termbox.Flush(); err != nil {
		return err
	}

	termbox.SetCursor(2, 1)
	if err := termbox.Flush(); err != nil {
		return err
	}

	time.Sleep(30 * time.Second)

	return nil
}

func drawBoard() {
	for i := 0; i < boardSize*4+1; i++ {
		for j := 0; j < boardSize*2+1; j++ {
			var value rune
			switch {
			case j%2 == 0 && i%4 == 0:
				value = '+'
			case j%2 == 0:
				value = '-'
			case i%4 == 0:
				value = '|'
			}
			termbox.SetCell(i, j, value, termbox.ColorMagenta, termbox.ColorDefault)
		}
	}
}
