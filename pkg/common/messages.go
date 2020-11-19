package common

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	HostGameAction      = "hostGame"
	StartSoloGameAction = "startSoloGame"
	JoinGameAction      = "joinGame"

	ListOpenGamesAction = "listOpenGames"
	OpenGamesAction     = "openGames"

	PlaceDiskAction   = "placeDisk"
	UpdateBoardAction = "updateBoard"

	ErrorAction = "error"
)

var actionToMessage = map[string]interface{}{
	HostGameAction:      HostGameMessage{},
	StartSoloGameAction: StartSoloGameMessage{},
	JoinGameAction:      JoinGameMessage{},

	ListOpenGamesAction: BaseMessage{},
	OpenGamesAction:     OpenGamesMessage{},

	PlaceDiskAction:   PlaceDiskMessage{},
	UpdateBoardAction: UpdateBoardMessage{},

	ErrorAction: ErrorMessage{},
}

const BoardSize = 8

type Disk uint8

const (
	Player1 = Disk(1)
	Player2 = Disk(2)
)

type Board [BoardSize][BoardSize]Disk

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

type HostGameMessage struct {
	Action   string `json:"action"`
	Nickname string `json:"host"`
}

func NewHostGameMessage(nickname string) HostGameMessage {
	return HostGameMessage{
		Action:   HostGameAction,
		Nickname: nickname,
	}
}

type StartSoloGameMessage struct {
	Action     string `json:"action"`
	Nickname   string `json:"nickname"`
	Difficulty int    `json:"difficulty"`
}

func NewStartSoloGameMessage(nickname string, difficulty int) StartSoloGameMessage {
	return StartSoloGameMessage{
		Action:     StartSoloGameAction,
		Nickname:   nickname,
		Difficulty: difficulty,
	}
}

type JoinGameMessage struct {
	Action   string `json:"action"`
	Nickname string `json:"nickname"`
	Host     string `json:"host"`
}

func NewJoinGameMessage(nickname, host string) JoinGameMessage {
	return JoinGameMessage{
		Action:   JoinGameAction,
		Nickname: nickname,
		Host:     host,
	}
}

func NewListOpenGamesMessage() BaseMessage {
	return BaseMessage{Action: ListOpenGamesAction}
}

type OpenGamesMessage struct {
	Action string   `json:"action"`
	Hosts  []string `json:"hosts"`
}

func NewOpenGamesMessage(hosts []string) OpenGamesMessage {
	return OpenGamesMessage{
		Action: OpenGamesAction,
		Hosts:  hosts,
	}
}

type PlaceDiskMessage struct {
	Action   string `json:"action"`
	Nickname string `json:"nickname"`
	Host     string `json:"host"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

func NewPlaceDiskMessage(nickname, host string, x, y int) PlaceDiskMessage {
	return PlaceDiskMessage{
		Action:   PlaceDiskAction,
		Nickname: nickname,
		Host:     host,
		X:        x,
		Y:        y,
	}
}

type UpdateBoardMessage struct {
	Action string `json:"action"`
	Board  Board  `json:"board"`
	Player Disk   `json:"player"`
}

func NewUpdateBoardMessage(board Board, player Disk) UpdateBoardMessage {
	return UpdateBoardMessage{
		Action: UpdateBoardAction,
		Board:  board,
		Player: player,
	}
}

type ErrorMessage struct {
	Action string `json:"action"`
	Error  string `json:"error"`
}

func NewErrorMessage(err string) ErrorMessage {
	return ErrorMessage{
		Action: ErrorAction,
		Error:  err,
	}
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
