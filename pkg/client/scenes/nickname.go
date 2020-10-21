package scenes

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/nsf/termbox-go"
)

const maxNicknameLen = 10

type Nickname struct {
	scene
	nickname       string
	changeNickname bool
}

func (n *Nickname) Setup(changeScene ChangeScene, sendMessage SendMessage) error {
	if err := n.scene.Setup(changeScene, sendMessage); err != nil {
		return err
	}

	if err := n.load(); err != nil {
		return err
	}

	if n.nickname != "" && !n.changeNickname {
		return n.ChangeScene(&Menu{nickname: n.nickname})
	}

	return nil
}

func (n *Nickname) OnTerminalEvent(event termbox.Event) error {
	// Handle change scene.
	if event.Key == termbox.KeyEnter {
		if n.nickname == "" {
			return nil
		}

		if err := n.save(); err != nil {
			return err
		}

		return n.ChangeScene(&Menu{nickname: n.nickname})
	}

	// Handle typing.

	if event.Key == termbox.KeyBackspace2 {
		if n.nickname == "" {
			return nil
		}
		n.nickname = n.nickname[:len(n.nickname)-1]
	}

	var setLastChar func(rune)

	if len(n.nickname) < maxNicknameLen {
		setLastChar = func(r rune) {
			n.nickname += string(r)
		}
	} else {
		setLastChar = func(r rune) {
			n.nickname = n.nickname[:maxNicknameLen-1] + string(r)
		}
	}

	if event.Key == termbox.KeySpace {
		setLastChar(' ')
		return nil
	}

	letter := getLetter(event.Ch)
	if letter == 0 {
		return nil
	}

	setLastChar(letter)

	return nil
}

func getLetter(ch rune) rune {
	if unicode.IsLetter(ch) {
		return unicode.ToUpper(ch)
	}

	return 0
}

func (n *Nickname) Draw() {
	drawGameBoyBorder()
	drawSplash()

	draw(offset(center, 0, 2), normal, "Enter your name:")

	var sb strings.Builder
	sb.WriteString(n.nickname)
	for i := len(n.nickname); i < maxNicknameLen; i++ {
		sb.WriteRune('_')
	}

	draw(offset(center, 0, 4), normal, sb.String())

	cursorX := min(len(n.nickname), maxNicknameLen-1) - maxNicknameLen/2
	setCursor(offset(center, cursorX, 4))
}

func (n *Nickname) HasFreeKeyboardInput() bool {
	return true
}

func (n *Nickname) load() error {
	configPath, err := n.configPath()
	if err != nil {
		return err
	}

	nicknameBytes, err := ioutil.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil
	}

	n.nickname = string(nicknameBytes)

	return err
}

func (n *Nickname) save() error {
	configPath, err := n.configPath()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, []byte(n.nickname), 0600)
}

func (n *Nickname) configPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dirPath := path.Join(homedir, ".othelgo")
	filePath := path.Join(dirPath, "nickname")

	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return "", err
	}

	return filePath, nil
}
