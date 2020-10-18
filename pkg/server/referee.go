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
		nextX := v[0] + x
		nextY := v[1] + y
		if nextX < 0 || nextX >= messages.BoardSize || nextY < 0 || nextY >= messages.BoardSize {
			continue
		}

		if board[nextX][nextY] == player%2+1 {
			// expand and aggregate vectors
			if ExpandVector(&board, nextX, nextY, player, v) {
				board[x][y] = player
				updated = true
			}
		}
	}
	return board, updated
}

func ExpandVector(board *messages.Board, x int, y int, player int, v [2]int) bool {
	// By the time ExpandVector is called, we have already chosen a vector from the position of the
	// placed disk that contains at least one of the opposing player's disks. Therefore, we need to
	// search along the vector for the next disk belonging to the current player.

	// board[x][y] = player
	// return true

	nextX := v[0] + x
	nextY := v[1] + y

	if nextX < 0 || nextX >= messages.BoardSize || nextY < 0 || nextY >= messages.BoardSize {
		return false
	}

	switch (*board)[nextX][nextY] {
	case 0:
		return false
	case player:
		(*board)[x][y] = player
		return true
	default:
		return ExpandVector(board, nextX, nextY, player, v)
	}
}
