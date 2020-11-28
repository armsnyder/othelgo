package testutil

import (
	"github.com/onsi/ginkgo"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Fail wraps ginkgo.Fail but also dumps an output of the Dynamo DB table items on failure.
func Fail(message string, callerSkip ...int) {
	dumpTable()
	ginkgo.Fail(message, callerSkip...)
}

func NewGameBoard() common.Board {
	return BuildBoard([]Move{{3, 3}, {4, 4}}, []Move{{3, 4}, {4, 3}})
}

type Move [2]int

func BuildBoard(p1, p2 []Move) (board common.Board) {
	for i, moves := range [][]Move{p1, p2} {
		player := common.Disk(i + 1)

		for _, move := range moves {
			x, y := move[0], move[1]
			board[x][y] = player
		}
	}

	return board
}

func Send(client **Client, messageToSend interface{}) func() {
	return func() {
		(*client).Send(messageToSend)
	}
}
