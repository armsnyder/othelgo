package scenes

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const (
	buttonNormal = iota
	buttonEasy
	buttonHard
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
			m.button = buttonNormal
		}
	case dx == 1:
		switch m.button {
		case buttonEasy, buttonNormal, buttonHard:
			m.button = buttonHostGame
		case buttonHostGame, buttonJoinGame:
			m.button = buttonChangeName
		}
	case dy == -1:
		switch m.button {
		case buttonNormal:
			m.button = buttonEasy
		case buttonHard:
			m.button = buttonNormal
		case buttonJoinGame:
			m.button = buttonHostGame
		default:
			m.button = buttonChangeName
		}
	case dy == 1:
		switch m.button {
		case buttonEasy:
			m.button = buttonNormal
		case buttonNormal:
			m.button = buttonHard
		case buttonHostGame:
			m.button = buttonJoinGame
		case buttonChangeName:
			m.button = buttonHostGame
		}
	}

	if event.Key == termbox.KeyEnter {
		switch m.button {
		case buttonEasy:
			return m.ChangeScene(&Game{player: 1, difficulty: 0, nickname: m.nickname})
		case buttonNormal:
			return m.ChangeScene(&Game{player: 1, difficulty: 1, nickname: m.nickname})
		case buttonHard:
			return m.ChangeScene(&Game{player: 1, difficulty: 2, nickname: m.nickname})
		case buttonHostGame:
			return m.ChangeScene(&Game{player: 1, multiplayer: true, nickname: m.nickname})
		case buttonJoinGame:
			return m.ChangeScene(&Game{player: 2, multiplayer: true, nickname: m.nickname})
		case buttonChangeName:
			return m.ChangeScene(&Nickname{ChangeNickname: true})
		}
	}

	return nil
}

func (m *Menu) Draw() {
	drawGameBoyBorder()
	drawSplash()

	draw(topRight, normal, fmt.Sprintf("Did you know? Your name is %s!", m.nickname))

	buttonColors := [6]color{normal, normal, normal, normal, normal, normal}
	buttonColors[m.button] = inverted

	multiplayerButtonColor := normal
	multiplayerOffset := offset(centerRight, 1, 3)
	if m.button == buttonHostGame || m.button == buttonJoinGame {
		multiplayerButtonColor = inverted
		draw(offset(multiplayerOffset, 1, 2), buttonColors[buttonHostGame], "[ HOST GAME ]")
		draw(offset(multiplayerOffset, 1, 4), buttonColors[buttonJoinGame], "[ JOIN GAME ]")
	}

	singleplayerButtonColor := normal
	singleplayerOffset := offset(centerLeft, -1, 3)
	if m.button == buttonEasy || m.button == buttonNormal || m.button == buttonHard {
		singleplayerButtonColor = inverted
		draw(offset(singleplayerOffset, -4, 2), buttonColors[buttonEasy], "[ EASY ]")
		draw(offset(singleplayerOffset, -3, 4), buttonColors[buttonNormal], "[ NORMAL ]")
		draw(offset(singleplayerOffset, -4, 6), buttonColors[buttonHard], "[ HARD ]")
	}

	draw(offset(centerLeft, -1, 3), singleplayerButtonColor, "[ SINGLEPLAYER ]")
	draw(multiplayerOffset, multiplayerButtonColor, "[ MULTIPLAYER ]")
	draw(offset(topRight, 0, 2), buttonColors[buttonChangeName], "[ CHANGE NAME ]")
}
