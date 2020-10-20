package client

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/client/scenes"

	"github.com/armsnyder/othelgo/pkg/common"
)

var firstScene = new(scenes.Nickname)

func Run() (err error) {
	// Setup log file.
	finish, err := setupFileLogger()
	if err != nil {
		return err
	}
	defer finish(err)

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

	// Setup a handler for changing scenes, and start the first scene.
	var currentScene scenes.Scene
	if err := setupChangeSceneHandler(&currentScene, c); err != nil {
		return err
	}

	// Listen for terminal events.
	terminalEvents := make(chan termbox.Event)
	go receiveTerminalEvents(terminalEvents)

	// Listen for websocket messages.
	messageQueue := make(chan common.AnyMessage)
	messageErrors := make(chan error)
	go receiveMessages(c, messageQueue, messageErrors)

	// Setup a ticker for calling Tick on the scene.
	ticker := time.NewTicker(time.Second / 12)
	defer ticker.Stop()

	// Run an event loop and call handlers on the current scene.
	for {
		select {
		case <-ticker.C:
			if currentScene.Tick() {
				if err := drawAndFlush(currentScene); err != nil {
					return err
				}
			}

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

func setupFileLogger() (finish func(err error), err error) {
	logFile, err := os.Create("othelgo.log")
	if err != nil {
		return nil, err
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.Lmsgprefix)

	finish = func(err error) {
		log.SetPrefix("")

		if err != nil {
			log.Printf("Exiting with error: %v", err)
		} else {
			log.Println("Exiting OK")
		}

		logFile.Close()

		// Reset logger.
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	}

	return finish, nil
}

func setupChangeSceneHandler(currentScene *scenes.Scene, c *websocket.Conn) error {
	sendMessage := func(v interface{}) error {
		action := reflect.ValueOf(v).FieldByName("Action").String()
		log.Printf("Sending message (action=%q)", action)

		return c.WriteJSON(v)
	}

	var changeScene scenes.ChangeScene
	changeScene = func(scene scenes.Scene) error {
		name := reflect.TypeOf(scene).Elem().Name()
		log.Printf("Changing scene to %s", name)
		log.SetPrefix(fmt.Sprintf("[%s] ", name))

		*currentScene = scene

		if err := scene.Setup(changeScene, sendMessage); err != nil {
			return err
		}

		return drawAndFlush(*currentScene)
	}

	return changeScene(firstScene)
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
