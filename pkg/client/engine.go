package client

import (
	"fmt"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/client/scenes"

	"github.com/armsnyder/othelgo/pkg/messages"
)

var allScenes = map[string]scenes.Scene{
	"menu": new(scenes.Menu),
	"game": new(scenes.Game),
}

const firstScene = "menu"

func Run() error {
	// Setup websocket.
	c, _, err := websocket.DefaultDialer.Dial("wss://1y9vcb5geb.execute-api.us-west-2.amazonaws.com/development", nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Setup terminal.
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	// Setup handler for changing scenes.

	var (
		currentScene scenes.Scene
		changeScene  scenes.ChangeScene
	)

	changeScene = func(name string, sceneContext scenes.SceneContext) error {
		nextScene, ok := allScenes[name]
		if !ok {
			return fmt.Errorf("no scene with name %q", name)
		}

		currentScene = nextScene

		if err := currentScene.Setup(changeScene, c.WriteJSON, sceneContext); err != nil {
			return err
		}

		return drawAndFlush(currentScene)
	}

	// Set the first scene.
	if err := changeScene(firstScene, nil); err != nil {
		return err
	}

	// Listen for terminal events.
	terminalEvents := make(chan termbox.Event)
	go receiveTerminalEvents(terminalEvents)

	// Listen for websocket messages.
	messageQueue := make(chan messages.AnyMessage)
	messageErrors := make(chan error)
	go receiveMessages(c, messageQueue, messageErrors)

	// Run an event loop and call handlers on the current scene.
	for {
		select {
		case event := <-terminalEvents:
			if shouldInterrupt(event) {
				termbox.Interrupt()
				return nil
			}

			if err := currentScene.OnTerminalEvent(event); err != nil {
				return err
			}

			if err := drawAndFlush(currentScene); err != nil {
				return err
			}

		case message := <-messageQueue:
			if err := currentScene.OnMessage(message); err != nil {
				return err
			}

			if err := drawAndFlush(currentScene); err != nil {
				return err
			}

		case err := <-messageErrors:
			return err
		}
	}
}

func receiveTerminalEvents(ch chan<- termbox.Event) {
	for {
		event := termbox.PollEvent()
		ch <- event
	}
}

func receiveMessages(c *websocket.Conn, messageQueue chan<- messages.AnyMessage, messageErrors chan<- error) {
	for {
		var message messages.AnyMessage
		if err := c.ReadJSON(&message); err != nil {
			messageErrors <- err
		}
		messageQueue <- message
	}
}

func shouldInterrupt(event termbox.Event) bool {
	return unicode.ToLower(event.Ch) == 'q' || event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyEsc
}

func drawAndFlush(scene scenes.Scene) error {
	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		return err
	}

	scene.Draw()

	return termbox.Flush()
}
