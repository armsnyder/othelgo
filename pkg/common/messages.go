package common

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	PlaceDiskAction   = "placeDisk"
	UpdateBoardAction = "updateBoard"
	NewGameAction     = "newGame"
	JoinGameAction    = "joinGame"
)

var actionToMessage = map[string]interface{}{
	PlaceDiskAction:   PlaceDiskMessage{},
	UpdateBoardAction: UpdateBoardMessage{},
	NewGameAction:     NewGameMessage{},
	JoinGameAction:    JoinGameMessage{},
}

const BoardSize = 8

type Board [BoardSize][BoardSize]int

func (b Board) String() string {
	// This function makes Board implement fmt.Stringer so that it renders visually in test outputs.
	var sb strings.Builder
	for y := 0; y < BoardSize; y++ {
		sb.WriteRune('\n')
		for x := 0; x < BoardSize; x++ {
			var ch rune
			switch b[x][y] {
			case 0:
				ch = '_'
			case 1:
				ch = 'x'
			case 2:
				ch = 'o'
			}
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}

type BaseMessage struct {
	Action string `json:"action"`
}

type PlaceDiskMessage struct {
	Action string `json:"action"`
	Player int    `json:"player"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

func NewPlaceDiskMessage(player, x, y int) PlaceDiskMessage {
	return PlaceDiskMessage{
		Action: PlaceDiskAction,
		Player: player,
		X:      x,
		Y:      y,
	}
}

type UpdateBoardMessage struct {
	Action string `json:"action"`
	Board  Board  `json:"board"`
	Player int    `json:"player"`
}

func NewUpdateBoardMessage(board Board, player int) UpdateBoardMessage {
	return UpdateBoardMessage{
		Action: UpdateBoardAction,
		Board:  board,
		Player: player,
	}
}

type NewGameMessage BaseMessage

func NewNewGameMessage() NewGameMessage {
	return NewGameMessage{Action: NewGameAction}
}

type JoinGameMessage BaseMessage

func NewJoinGameMessage() JoinGameMessage {
	return JoinGameMessage{Action: JoinGameAction}
}

type AnyMessage struct {
	Action  string
	Message interface{}
}

func (u *AnyMessage) UnmarshalJSON(data []byte) error {
	var base BaseMessage
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}

	if base.Action == "" {
		return fmt.Errorf("invalid message %q", string(data))
	}

	prototype, ok := actionToMessage[base.Action]
	if !ok {
		return fmt.Errorf("unsupported message action %q", base.Action)
	}

	message := reflect.New(reflect.TypeOf(prototype)).Interface()

	if err := json.Unmarshal(data, message); err != nil {
		return err
	}

	u.Action = base.Action
	u.Message = message

	return nil
}
