package client

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/client/scenes"

	"github.com/armsnyder/othelgo/pkg/common"
)

var allScenes = map[string]scenes.Scene{
	"menu": new(scenes.Menu),
	"game": new(scenes.Game),
}

const firstScene = "menu"

func Run() (err error) {
	// Setup log file.

	logFile, err := os.Create("othelgo.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.Lmsgprefix)

	defer func() {
		log.SetPrefix("")

		if err != nil {
			log.Printf("Exiting with error: %v", err)
		} else {
			log.Println("Exiting OK")
		}

		// Reset logger.
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	}()

	// Setup websocket.
	addr := "wss://1y9vcb5geb.execute-api.us-west-2.amazonaws.com/development"
	log.Printf("Dialing websocket %q", addr)

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return err
	}

	defer c.Close()

	// Setup terminal.

	log.Println("Initializing terminal")

	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	// Setup scene handlers.

	sendMessage := func(v interface{}) error {
		action := reflect.ValueOf(v).FieldByName("Action").String()
		log.Printf("Sending message (action=%q)", action)

		return c.WriteJSON(v)
	}

	var (
		currentScene scenes.Scene
		changeScene  scenes.ChangeScene
	)

	changeScene = func(name string, sceneContext scenes.SceneContext) error {
		log.Printf("Changing scene to %q", name)

		nextScene, ok := allScenes[name]
		if !ok {
			return fmt.Errorf("no scene with name %q", name)
		}

		currentScene = nextScene

		log.SetPrefix(fmt.Sprintf("[%s] ", name))

		if err := currentScene.Setup(changeScene, sendMessage, sceneContext); err != nil {
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

	messageQueue := make(chan common.AnyMessage)
	messageErrors := make(chan error)

	go receiveMessages(c, messageQueue, messageErrors)

	// Run an event loop and call handlers on the current scene.
	for {
		select {
		case event := <-terminalEvents:
			log.Printf("Received terminal event (type=%d)", event.Type)

			if shouldInterrupt(event) {
				log.Println("Interrupting terminal")
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
			log.Printf("Received message (action=%q)", message.Action)

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

func receiveMessages(c *websocket.Conn, messageQueue chan<- common.AnyMessage, messageErrors chan<- error) {
	for {
		var message common.AnyMessage
		if err := c.ReadJSON(&message); err != nil {
			messageErrors <- fmt.Errorf("failed to read message from websocket: %w", err)
		}
		messageQueue <- message
	}
}

func shouldInterrupt(event termbox.Event) bool {
	return unicode.ToLower(event.Ch) == 'q' || event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyEsc
}

func drawAndFlush(scene scenes.Scene) error {
	log.Println("Drawing")

	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		return err
	}

	scene.Draw()

	return termbox.Flush()
}
