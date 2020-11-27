package scenes

import (
	"fmt"
	"unicode"

	"github.com/armsnyder/othelgo/pkg/client/draw"
	"github.com/armsnyder/othelgo/pkg/messages"

	"github.com/nsf/termbox-go"
)

type Join struct {
	scene
	nickname string
	hosts    []string
	selected int
}

func (j *Join) Setup(changeScene ChangeScene, sendMessage SendMessage) error {
	if err := j.scene.Setup(changeScene, sendMessage); err != nil {
		return err
	}

	return sendMessage(messages.ListOpenGames{})
}

func (j *Join) OnMessage(message interface{}) error {
	if m, ok := message.(*messages.OpenGames); ok {
		j.hosts = m.Hosts
	}
	if len(j.hosts) > 0 {
		j.selected = 0
	}

	return nil
}

func (j *Join) OnTerminalEvent(event termbox.Event) error {
	if event.Key == termbox.KeyEnter && len(j.hosts) > 0 {
		return j.ChangeScene(&Game{player: 2, multiplayer: true, nickname: j.nickname, host: j.hosts[j.selected]})
	}
	_, dy := getDirectionPressed(event)
	switch {
	case dy == -1 && j.selected > 0:
		j.selected--
	case dy == 1 && j.selected < len(j.hosts)-1:
		j.selected++
	}

	if unicode.ToUpper(event.Ch) == 'M' {
		return j.ChangeScene(&Menu{nickname: j.nickname})
	}
	return nil
}

func (j *Join) Draw() {
	draw.Draw(draw.TopRight, draw.Normal, fmt.Sprintf("Your name is %s!", j.nickname))
	draw.Draw(draw.BotRight, draw.Normal, "[M] MENU  [Q] QUIT")

	if len(j.hosts) > 0 {
		buttonColors := [6]draw.Color{}
		for i := range buttonColors {
			buttonColors[i] = draw.Normal
		}
		buttonColors[j.selected] = draw.Inverted
		draw.Draw(draw.Offset(draw.CenterRight, -9, 0), draw.Normal, "=== OPEN GAMES ===")
		for i, h := range j.hosts {
			os := -(len(h) + 4) / 2
			draw.Draw(draw.Offset(draw.CenterRight, os, i*2+2), buttonColors[i], fmt.Sprintf("[ %s ]", h))
		}
	} else {
		draw.Draw(draw.CenterTop, draw.Normal, "MORE LIKE \"NO GAME\"")
	}
}
