package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
)

const boardSize = 8

func main() {
	c, res, err := websocket.DefaultDialer.Dial("wss://1y9vcb5geb.execute-api.us-west-2.amazonaws.com/development", nil) // returns conn, HTTPres, err
	log.Println(res)
	str, _ := ioutil.ReadAll(res.Body)
	log.Println(string(str))
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	drawBoard()

	if err := termbox.Flush(); err != nil {
		log.Fatal(err)
	}

	termbox.SetCursor(2, 1)
	if err := termbox.Flush(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(30 * time.Second)
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
