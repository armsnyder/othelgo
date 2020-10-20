package scenes

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const (
	buttonNewGame = iota
	buttonJoinGame
	buttonChangeName
)

type Menu struct {
	scene
	button   int
	nickname string
}

func (m *Menu) OnTerminalEvent(event termbox.Event) error {
	dx, dy := getDirectionPressed(event)

	switch {
	case dx == -1:
		switch m.button {
		case buttonChangeName:
			m.button = buttonJoinGame
		case buttonJoinGame:
			m.button = buttonNewGame
		}
	case dx == 1:
		switch m.button {
		case buttonNewGame:
			m.button = buttonJoinGame
		case buttonJoinGame:
			m.button = buttonChangeName
		}
	case dy == -1:
		m.button = buttonChangeName
	case dy == 1:
		if m.button == buttonChangeName {
			m.button = buttonJoinGame
		}
	}

	if event.Key == termbox.KeyEnter {
		switch m.button {
		case buttonNewGame:
			return m.ChangeScene(&Game{player: 1})
		case buttonJoinGame:
			return m.ChangeScene(&Game{player: 2})
		case buttonChangeName:
			return m.ChangeScene(&Nickname{changeNickname: true, nickname: m.nickname})
		}
	}

	return nil
}

func (m *Menu) Draw() {
	drawGameBoyBorder()
	drawSplash()

	draw(topRight, normal, fmt.Sprintf("Did you know? Your name is %s!", m.nickname))

	buttonColors := [3]color{normal, normal, normal}
	buttonColors[m.button] = inverted

	draw(offset(centerLeft, -1, 3), buttonColors[buttonNewGame], "[ NEW GAME ]")
	draw(offset(centerRight, 1, 3), buttonColors[buttonJoinGame], "[ JOIN GAME ]")
	draw(offset(topRight, 0, 2), buttonColors[buttonChangeName], "[ CHANGE NAME ]")
}
