package client

import (
	"fmt"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/messages"
)

var allScenes = map[string]scene{
	"game": gameScene,
}

const firstScene = "game"

// scene is responsible for the logic and view of a particular page of the application.
// It can handle websocket messages and terminal events.
type scene func(changeScene func(string), sendMessage func(interface{}), onError func(error)) (onMessage func(messages.AnyMessage), onEvent func(termbox.Event))

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

	// Setup handlers for scenes to use.

	var (
		changeScene func(nextScene string)
		sendMessage func(interface{})
		onError     func(error)
		onMessage   func(messages.AnyMessage)
		onEvent     func(termbox.Event)
	)

	errors := make(chan error)

	onError = func(err error) {
		errors <- err
	}

	changeScene = func(name string) {
		if nextScene, ok := allScenes[name]; ok {
			onMessage, onEvent = nextScene(changeScene, sendMessage, onError)
		} else {
			onError(fmt.Errorf("no scene with name %q", name))
		}
	}

	sendMessage = func(v interface{}) {
		if err := c.WriteJSON(v); err != nil {
			onError(err)
		}
	}

	// Set the first scene.
	changeScene(firstScene)

	// Listen for terminal events.
	terminalEvents := make(chan termbox.Event)
	go receiveTerminalEvents(terminalEvents)

	// Listen for websocket messages.
	messageQueue := make(chan messages.AnyMessage)
	go receiveMessages(c, messageQueue, errors)

	// Run an event loop and call the current handlers configured by the current scene.
	for {
		select {
		case event := <-terminalEvents:
			if shouldInterrupt(event) {
				termbox.Interrupt()
				return nil
			}
			onEvent(event)

		case anyMessage := <-messageQueue:
			onMessage(anyMessage)

		case err := <-errors:
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
