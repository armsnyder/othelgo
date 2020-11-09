package scenes

import (
	"fmt"
	"unicode"

	"github.com/armsnyder/othelgo/pkg/common"

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

	return sendMessage(common.NewListOpenGamesMessage())
}

func (j *Join) OnMessage(message common.AnyMessage) error {
	if m, ok := message.Message.(*common.OpenGamesMessage); ok {
		j.hosts = m.Hosts
	}
	if len(j.hosts) > 0 {
		j.selected = 0
	}

	return nil
}

func (j *Join) OnTerminalEvent(event termbox.Event) error {
	if event.Key == termbox.KeyEnter && len(j.hosts) > 0 {
		return j.ChangeScene(&Game{player: 2, multiplayer: true, nickname: j.nickname, opponent: j.hosts[j.selected]})
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
	drawGameBoyBorder()
	draw(topRight, normal, fmt.Sprintf("Your name is %s!", j.nickname))

	if len(j.hosts) > 0 {
		buttonColors := [6]color{}
		for i := range buttonColors {
			buttonColors[i] = normal
		}
		buttonColors[j.selected] = inverted
		draw(topLeft, normal, "OPEN GAMES")
		for i, h := range j.hosts {
			draw(offset(topLeft, 0, i*2+2), buttonColors[i], fmt.Sprintf("[ %s ]", h))
		}
	} else {
		draw(centerTop, normal, "MORE LIKE \"NO GAME\"")
	}
}
