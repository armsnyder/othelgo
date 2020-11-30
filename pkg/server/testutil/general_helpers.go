package testutil

import (
	"github.com/onsi/ginkgo"

	"github.com/armsnyder/othelgo/pkg/common"
)

// DumpTableOnFailure can be passed to ginkgo.JustAfterEach to dump the DB table to log output
// if an assertion failed.
func DumpTableOnFailure() {
	if ginkgo.CurrentGinkgoTestDescription().Failed {
		dumpTable()
	}
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

// Send is a convenience wrapper around Client.Send which can be used directly as an argument to
// ginkgo.BeforeEach.
func Send(client **Client, messageToSend interface{}) func() {
	return func() {
		(*client).Send(messageToSend)
	}
}
