package testutil

import (
	"strings"

	. "github.com/onsi/gomega" //nolint:golint

	"github.com/armsnyder/othelgo/pkg/common"
	"github.com/armsnyder/othelgo/pkg/messages"
)

// Common assertions that can be passed to ginkgo.It. Generally, assertions can be generalized here
// if they are more then two lines of code and are used more then three times throughout tests.

func ExpectNewGameBoard(client **Client) func() {
	newGameBoard := BuildBoard([]Move{{3, 3}, {4, 4}}, []Move{{3, 4}, {4, 3}})
	return func() {
		var message messages.UpdateBoard
		Expect(*client).To(HaveReceived(&message))
		Expect(message.Board).To(Equal(newGameBoard))
	}
}

func ExpectNoOpenGames(client **Client) func() {
	return ExpectOpenGames(client)
}

func ExpectOpenGames(client **Client, hosts ...string) func() {
	var hostInterfaces []interface{}
	for _, host := range hosts {
		hostInterfaces = append(hostInterfaces, host)
	}

	return func() {
		var message messages.OpenGames
		Expect(*client).To(HaveReceived(&message))
		Expect(message.Hosts).To(ConsistOf(hostInterfaces...))
	}
}

func ExpectPlayerLeft(client **Client, player string) func() {
	return func() {
		var message messages.GameOver
		Expect(*client).To(HaveReceived(&message))
		Expect(message.Message).To(Equal(strings.ToUpper(player) + " left the game"))
	}
}

func ExpectTurn(client **Client, player common.Disk) func() {
	return func() {
		var message messages.UpdateBoard
		Expect(*client).To(HaveReceived(&message))
		Expect(message.Player).To(Equal(player))
	}
}
