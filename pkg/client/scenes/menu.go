package scenes

import (
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
	drawUpperRight("Did you know? Your name is " + m.nickname + "!")
	drawFromCenter(splashText, 0, -6, termbox.ColorDefault, termbox.ColorDefault)

	var (
		newGameText  = "[ NEW GAME ]"
		joinGameText = "[ JOIN GAME ]"
	)

	drawStringHighlight := func(s string, dx, dy int, highlight bool) {
		var fg, bg termbox.Attribute
		if highlight {
			fg = termbox.ColorBlack
			bg = termbox.ColorWhite
		}

		drawFromCenter(s, dx, dy, fg, bg)
	}

	drawStringHighlight(newGameText, -len(newGameText)/2-1, 3, !m.isJoinGame)
	drawStringHighlight(joinGameText, len(joinGameText)/2+1, 3, m.isJoinGame)
}
