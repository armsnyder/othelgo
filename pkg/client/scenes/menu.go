package scenes

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const (
	buttonSingleplayer = iota
	buttonHostGame
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
			m.button = buttonHostGame
		case buttonHostGame, buttonJoinGame:
			m.button = buttonSingleplayer
		}
	case dx == 1:
		switch m.button {
		case buttonSingleplayer:
			m.button = buttonHostGame
		case buttonHostGame, buttonJoinGame:
			m.button = buttonChangeName
		}
	case dy == -1:
		switch m.button {
		case buttonJoinGame:
			m.button = buttonHostGame
		default:
			m.button = buttonChangeName
		}
	case dy == 1:
		switch m.button {
		case buttonHostGame:
			m.button = buttonJoinGame
		case buttonChangeName:
			m.button = buttonHostGame
		}
	}

	if event.Key == termbox.KeyEnter {
		switch m.button {
		case buttonSingleplayer:
			return m.ChangeScene(&Game{player: 1, multiplayer: false, nickname: m.nickname})
		case buttonHostGame:
			return m.ChangeScene(&Game{player: 1, multiplayer: true, nickname: m.nickname})
		case buttonJoinGame:
			return m.ChangeScene(&Game{player: 2, multiplayer: true, nickname: m.nickname})
		case buttonChangeName:
			return m.ChangeScene(&Nickname{changeNickname: true})
		}
	}

	return nil
}

func (m *Menu) Draw() {
	drawGameBoyBorder()
	drawSplash()

	draw(topRight, normal, fmt.Sprintf("Did you know? Your name is %s!", m.nickname))

	buttonColors := [4]color{normal, normal, normal, normal}
	buttonColors[m.button] = inverted

	multiplayerButtonColor := normal
	multiplayerOffset := offset(centerRight, 1, 3)
	if m.button == buttonHostGame || m.button == buttonJoinGame {
		multiplayerButtonColor = inverted
		draw(offset(multiplayerOffset, 0, 2), buttonColors[buttonHostGame], "[ HOST GAME ]")
		draw(offset(multiplayerOffset, 0, 4), buttonColors[buttonJoinGame], "[ JOIN GAME ]")
	}

	draw(offset(centerLeft, -1, 3), buttonColors[buttonSingleplayer], "[ SINGLEPLAYER ]")
	draw(multiplayerOffset, multiplayerButtonColor, "[ MULTIPLAYER ]")
	draw(offset(topRight, 0, 2), buttonColors[buttonChangeName], "[ CHANGE NAME ]")
}
