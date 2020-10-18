package server

import (
	"github.com/armsnyder/othelgo/pkg/messages"
)

func ApplyMove(board messages.Board, x int, y int, player int) (messages.Board, bool) {
	// validate x and y are on the even board
	if x < 0 || x >= messages.BoardSize || y < 0 || y >= messages.BoardSize {
		return board, false
	}
	// verify cell is empty
	if board[x][y] != 0 {
		return board, false
	}
	// choose vectors
	updated := false
	for _, v := range [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}} {
		if v[0]+x < 0 || v[0]+x >= messages.BoardSize || v[1]+y < 0 || v[1]+y >= messages.BoardSize {
			continue
		}
		if board[v[0]+x][v[1]+y] == player%2+1 {
			// expand and aggregate vectors
			if ExpandVector(&board, x, y, player, v) {
				updated = true
			}
		}
	}
	return board, updated
}

func ExpandVector(board *messages.Board, x int, y int, player int, v [2]int) bool {
	board[x][y] = player
	return true
}
