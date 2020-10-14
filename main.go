package main

import (
	"time"
	"log"

	"github.com/nsf/termbox-go"
)

const boardSize = 8

func main() {
	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	drawBoard()

	if err := termbox.Flush(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(5*time.Second)
}

func drawBoard() {
	for i := 0; i < boardSize*2+1; i++ {
		for j := 0; j < boardSize*2+1; j++ {
			var value rune
			switch {
			case j%2 == 0 && i%2 == 0:
				value = '+'
			case j%2 == 0:
				value = '-'
			case i%2 == 0:
				value = '|'
			}
			termbox.SetCell(i, j, value, termbox.ColorMagenta, termbox.ColorDefault)
		}
	}
}
