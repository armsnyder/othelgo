package testutil

import (
	"github.com/armsnyder/othelgo/pkg/messages"
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

func MakeaBunchaMoves(host string, player1 *Client, player1Name string, player2 *Client, player2Name string, moves []Move) {
	players := []*Client{player1, player2}
	playerNames := []string{player1Name, player2Name}
	for i, m := range moves {
		player := players[i%2]
		name := playerNames[i%2]
		player.Send(messages.PlaceDisk{Nickname: name, Host: host, X: m[0], Y: m[1]})
	}
}
