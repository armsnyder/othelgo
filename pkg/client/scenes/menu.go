package scenes

import (
	"fmt"

	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/client/draw"
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
			return m.ChangeScene(&Game{player: 1, difficulty: 0, nickname: m.nickname, host: m.nickname})
		case buttonNormal:
			return m.ChangeScene(&Game{player: 1, difficulty: 1, nickname: m.nickname, host: m.nickname})
		case buttonHard:
			return m.ChangeScene(&Game{player: 1, difficulty: 2, nickname: m.nickname, host: m.nickname})
		case buttonHostGame:
			return m.ChangeScene(&Game{player: 1, multiplayer: true, nickname: m.nickname, host: m.nickname})
		case buttonJoinGame:
			// return m.ChangeScene(&Game{player: 2, multiplayer: true, nickname: m.nickname})
			return m.ChangeScene(&Join{nickname: m.nickname})
		case buttonChangeName:
			return m.ChangeScene(&Nickname{ChangeNickname: true})
		}
	}

	return nil
}

func (m *Menu) Draw() {
	drawSplash()

	draw.Draw(draw.TopRight, draw.Normal, fmt.Sprintf("Did you know? Your name is %s!", m.nickname))

	buttonColors := [6]draw.Color{draw.Normal, draw.Normal, draw.Normal, draw.Normal, draw.Normal, draw.Normal}
	buttonColors[m.button] = draw.Inverted

	multiplayerButtonColor := draw.Normal
	multiplayerOffset := draw.Offset(draw.CenterRight, 1, 3)
	if m.button == buttonHostGame || m.button == buttonJoinGame {
		multiplayerButtonColor = draw.Inverted
		draw.Draw(draw.Offset(multiplayerOffset, 1, 2), buttonColors[buttonHostGame], "[ HOST GAME ]")
		draw.Draw(draw.Offset(multiplayerOffset, 1, 4), buttonColors[buttonJoinGame], "[ JOIN GAME ]")
	}

	singleplayerButtonColor := draw.Normal
	singleplayerOffset := draw.Offset(draw.CenterLeft, -1, 3)
	if m.button == buttonEasy || m.button == buttonNormal || m.button == buttonHard {
		singleplayerButtonColor = draw.Inverted
		draw.Draw(draw.Offset(singleplayerOffset, -4, 2), buttonColors[buttonEasy], "[ EASY ]")
		draw.Draw(draw.Offset(singleplayerOffset, -3, 4), buttonColors[buttonNormal], "[ NORMAL ]")
		draw.Draw(draw.Offset(singleplayerOffset, -4, 6), buttonColors[buttonHard], "[ HARD ]")
	}

	draw.Draw(draw.Offset(draw.CenterLeft, -1, 3), singleplayerButtonColor, "[ SINGLEPLAYER ]")
	draw.Draw(multiplayerOffset, multiplayerButtonColor, "[ MULTIPLAYER ]")
	draw.Draw(draw.Offset(draw.TopRight, 0, 2), buttonColors[buttonChangeName], "[ CHANGE NAME ]")
}
