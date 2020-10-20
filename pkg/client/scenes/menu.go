package scenes

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

type Menu struct {
	scene
	isJoinGame bool
	nickname   string
}

func (m *Menu) OnTerminalEvent(event termbox.Event) error {
	switch dx, _ := getDirectionPressed(event); dx {
	case -1:
		m.isJoinGame = false
	case 1:
		m.isJoinGame = true
	}

	if event.Key == termbox.KeyEnter {
		var player int
		if m.isJoinGame {
			player = 2
		} else {
			player = 1
		}

		return m.ChangeScene(&Game{player: player})
	}

	return nil
}

func (m *Menu) Draw() {
	drawGameBoyBorder()
	drawSplash()

	draw(topRight, normal, fmt.Sprintf("Did you know? Your name is %s!", m.nickname))

	var (
		newGameText  = "[ NEW GAME ]"
		joinGameText = "[ JOIN GAME ]"
	)

	var buttonColors [2]color
	if m.isJoinGame {
		buttonColors = [2]color{normal, inverted}
	} else {
		buttonColors = [2]color{inverted, normal}
	}

	draw(offset(centerLeft, -1, 3), buttonColors[0], newGameText)
	draw(offset(centerRight, 1, 3), buttonColors[1], joinGameText)
}
