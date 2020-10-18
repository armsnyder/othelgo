package scenes

import (
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/messages"
)

// Scene is responsible for the logic and view of a particular page of the application.
// It can handle websocket messages and terminal events.
type Scene interface {
	Setup(changeScene ChangeScene, sendMessage SendMessage, sceneContext SceneContext)
	OnMessage(message messages.AnyMessage) error
	OnTerminalEvent(event termbox.Event) error
	Draw()
}

// types for Scene setup method.
type (
	ChangeScene  func(string, SceneContext) error
	SendMessage  func(interface{}) error
	SceneContext map[string]interface{}
)

// scene has default implementations for Scene for convenience.
type scene struct {
	ChangeScene
	SendMessage
	SceneContext
}

func (b *scene) Setup(changeScene ChangeScene, sendMessage SendMessage, sceneContext SceneContext) {
	b.ChangeScene = changeScene
	b.SendMessage = sendMessage
	b.SceneContext = sceneContext
}

func (b *scene) OnMessage(_ messages.AnyMessage) error {
	// Default implementation is a no-op.
	return nil
}
