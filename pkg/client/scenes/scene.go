package scenes

import (
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Scene is responsible for the logic and view of a particular page of the application.
// It can handle websocket messages and terminal events.
type Scene interface {
	Setup(changeScene ChangeScene, sendMessage SendMessage) error
	OnMessage(message common.AnyMessage) error
	OnTerminalEvent(event termbox.Event) error
	Tick() bool
	Draw()
	HasFreeKeyboardInput() bool
	OnQuit()
}

// types for Scene setup method.
type (
	ChangeScene func(Scene) error
	SendMessage func(interface{}) error
)

// scene has default implementations for Scene for convenience.
type scene struct {
	ChangeScene
	SendMessage
}

func (s *scene) Setup(changeScene ChangeScene, sendMessage SendMessage) error {
	s.ChangeScene = changeScene
	s.SendMessage = sendMessage

	return nil
}

func (s *scene) OnMessage(_ common.AnyMessage) error {
	// Default implementation is a no-op.
	return nil
}

func (s *scene) Tick() bool {
	// Default implementation is a no-op.
	return false
}

func (s *scene) HasFreeKeyboardInput() bool {
	return false
}

func (s *scene) OnQuit() {
	// Default implementation is a no-op.
}
