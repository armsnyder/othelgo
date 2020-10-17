package messages

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	PlaceDiskAction   = "placeDisk"
	UpdateBoardAction = "updateBoard"
)

var actionToMessage = map[string]interface{}{
	PlaceDiskAction:   PlaceDiskMessage{},
	UpdateBoardAction: UpdateBoardMessage{},
}

const BoardSize = 8

type Board [BoardSize][BoardSize]int

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
}

func NewUpdateBoardMessage(board Board) UpdateBoardMessage {
	return UpdateBoardMessage{
		Action: UpdateBoardAction,
		Board:  board,
	}
}

type AnyMessage struct {
	Message interface{}
}

func (u *AnyMessage) UnmarshalJSON(data []byte) error {
	var base BaseMessage
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}

	prototype, ok := actionToMessage[base.Action]
	if !ok {
		return fmt.Errorf("unsupported message action %q", base.Action)
	}

	message := reflect.New(reflect.TypeOf(prototype)).Interface()

	if err := json.Unmarshal(data, message); err != nil {
		return err
	}

	u.Message = message

	return nil
}
