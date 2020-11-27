package messages

import "github.com/armsnyder/othelgo/pkg/common"

// To add a new message type, declare a new struct in this file and add it to the manifest variable.

// manifest must contain all message types.
var manifest = []interface{}{
	(*Hello)(nil),
	(*HostGame)(nil),
	(*StartSoloGame)(nil),
	(*JoinGame)(nil),
	(*Joined)(nil),
	(*LeaveGame)(nil),
	(*GameOver)(nil),
	(*ListOpenGames)(nil),
	(*OpenGames)(nil),
	(*PlaceDisk)(nil),
	(*UpdateBoard)(nil),
	(*Error)(nil),
	(*Decorate)(nil),
}

type Hello struct {
	Version string `json:"version"`
}

type HostGame struct {
	Nickname string `json:"nickname"`
}

type StartSoloGame struct {
	Nickname   string `json:"nickname"`
	Difficulty int    `json:"difficulty"`
}

type JoinGame struct {
	Nickname string `json:"nickname"`
	Host     string `json:"host"`
}

type Joined struct {
	Nickname string `json:"nickname"`
}

type LeaveGame struct {
	Nickname string `json:"nickname"`
	Host     string `json:"host"`
}

type GameOver struct {
	Message string `json:"message"`
}

type ListOpenGames struct{}

type OpenGames struct {
	Hosts []string `json:"hosts"`
}

type PlaceDisk struct {
	Nickname string `json:"nickname"`
	Host     string `json:"host"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type UpdateBoard struct {
	Board  common.Board `json:"board"`
	Player common.Disk  `json:"player"`
}

type Error struct {
	Error string `json:"error"`
}

type Decorate struct {
	Decoration string `json:"decoration"`
}
