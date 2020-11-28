package client

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"

	"github.com/armsnyder/othelgo/pkg/client/draw"
	"github.com/armsnyder/othelgo/pkg/client/scenes"

	"github.com/armsnyder/othelgo/pkg/messages"
)

func Run(local bool, version string) (err error) {
	// Setup log file.
	finish, err := setupFileLogger()
	if err != nil {
		return err
	}
	defer finish(err)

	// Setup websocket.
	c, finish2, err := setupWebsocket(local, version)
	if err != nil {
		return err
	}
	defer finish2()

	// Setup terminal.
	log.Println("Initializing terminal")
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	// Setup a handler for changing scenes, and start the first scene.
	var currentScene scenes.Scene
	var gameBorderDecoration string
	// We always want to prompt for a nickname when running locally because there will be more than
	// one client.
	firstScene := &scenes.Nickname{ChangeNickname: local}
	drawAndFlush := func() error { return drawAndFlushScene(currentScene, gameBorderDecoration) }
	if err := setupChangeSceneHandler(&currentScene, firstScene, drawAndFlush, c); err != nil {
		return err
	}

	// Listen for terminal events.
	terminalEvents := make(chan termbox.Event)
	go receiveTerminalEvents(terminalEvents)

	// Listen for websocket messages.
	messageQueue := make(chan interface{})
	messageErrors := make(chan error)
	go receiveMessages(c, messageQueue, messageErrors)

	// Setup a ticker for calling Tick on the scene.
	ticker := time.NewTicker(time.Second / 12)
	defer ticker.Stop()

	// Run an event loop and call handlers on the current scene.
	for {
		select {
		case <-ticker.C:
			if err := handleTick(currentScene, drawAndFlush); err != nil {
				return err
			}

		case event := <-terminalEvents:
			if err := handleTerminalEvent(event, currentScene, drawAndFlush); err != nil {
				return err
			}

		case message := <-messageQueue:
			if err := handleMessage(message, func(decoration string) { gameBorderDecoration = decoration }, currentScene, drawAndFlush); err != nil {
				return err
			}

		case err := <-messageErrors:
			log.Printf("error reading message from server: %v", err)
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

func setupWebsocket(local bool, version string) (*websocket.Conn, func(), error) {
	addr := "wss://1y9vcb5geb.execute-api.us-west-2.amazonaws.com/development"
	if local {
		addr = "ws://127.0.0.1:9000"
	}
	log.Printf("Dialing websocket %q", addr)
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, nil, err
	}
	err = c.WriteJSON(messages.Wrapper{Message: messages.Hello{Version: version}})
	if err != nil {
		return nil, nil, err
	}

	// Ping the server regularly to keep the connection open.
	go func() {
		for {
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to ping server: %v", err)
			}
			time.Sleep(time.Minute)
		}
	}()

	return c, func() { c.Close() }, nil
}

func setupChangeSceneHandler(currentScene *scenes.Scene, firstScene scenes.Scene, drawAndFlush func() error, c *websocket.Conn) error {
	sendMessage := func(v interface{}) error {
		log.Printf("Sending message %T", v)
		return c.WriteJSON(messages.Wrapper{Message: v})
	}

	var changeScene scenes.ChangeScene
	changeScene = func(scene scenes.Scene) error {
		name := reflect.TypeOf(scene).Elem().Name()
		log.Printf("Changing scene to %s", name)
		log.SetPrefix(fmt.Sprintf("[%s] ", name))

		*currentScene = scene

		termbox.HideCursor()

		if err := scene.Setup(changeScene, sendMessage); err != nil {
			return err
		}

		return drawAndFlush()
	}

	return changeScene(firstScene)
}

func receiveTerminalEvents(ch chan<- termbox.Event) {
	for {
		event := termbox.PollEvent()
		ch <- event
	}
}

func receiveMessages(c *websocket.Conn, messageQueue chan<- interface{}, messageErrors chan<- error) {
	for {
		var wrapper messages.Wrapper
		if err := c.ReadJSON(&wrapper); err != nil {
			messageErrors <- fmt.Errorf("failed to read message from websocket: %w", err)
		}
		messageQueue <- wrapper.Message
	}
}

func shouldInterrupt(event termbox.Event, scene scenes.Scene) bool {
	if unicode.ToLower(event.Ch) == 'q' && !scene.HasFreeKeyboardInput() {
		return true
	}

	return event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyEsc
}

func drawAndFlushScene(scene scenes.Scene, decoration string) error {
	log.Println("Drawing")

	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		return err
	}

	scene.Draw()

	draw.Border(decoration)

	return termbox.Flush()
}

func handleTick(currentScene scenes.Scene, drawAndFlush func() error) error {
	if currentScene.Tick() {
		if err := drawAndFlush(); err != nil {
			return err
		}
	}
	return nil
}

func handleTerminalEvent(event termbox.Event, currentScene scenes.Scene, drawAndFlush func() error) error {
	log.Printf("Received terminal event (type=%d)", event.Type)

	if shouldInterrupt(event, currentScene) {
		log.Println("Quitting scene")
		currentScene.OnQuit()

		log.Println("Interrupting terminal")
		termbox.Interrupt()

		return errors.New("interrupt")
	}

	if err := currentScene.OnTerminalEvent(event); err != nil {
		return err
	}

	return drawAndFlush()
}

func handleMessage(message interface{}, changeGameBorderDecoration func(string), currentScene scenes.Scene, drawAndFlush func() error) error {
	log.Printf("Received message %T", message)

	if m, ok := message.(*messages.Decorate); ok {
		changeGameBorderDecoration(m.Decoration)
	}

	if err := currentScene.OnMessage(message); err != nil {
		return err
	}

	if err := drawAndFlush(); err != nil {
		return err
	}
	return nil
}
