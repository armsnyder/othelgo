package scenes

import (
	"github.com/nsf/termbox-go"
)

type Menu struct {
	scene
	isJoinGame bool
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

		return m.ChangeScene("game", SceneContext{"player": player})
	}

	return nil
}

func (m *Menu) Draw() {
	var (
		titleText    = "OTHELGO"
		newGameText  = "[ NEW GAME ]"
		joinGameText = "[ JOIN GAME ]"
	)

	drawStringHighlight := func(s string, x, y int, highlight bool) {
		var fg, bg termbox.Attribute
		if highlight {
			fg = termbox.ColorWhite
			bg = termbox.ColorGreen
		}

		drawString(s, x, y, fg, bg)
	}

	drawStringHighlight(newGameText, 1, 2, !m.isJoinGame)
	drawStringHighlight(joinGameText, len(newGameText)+3, 2, m.isJoinGame)

	titleX := 2 + len(newGameText) - len(titleText)/2
	drawStringDefault(titleText, titleX, 0)
}
