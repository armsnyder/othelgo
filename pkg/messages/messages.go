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
	Version string `json:"version" validate:"semver"`
}

type HostGame struct {
	Nickname string `json:"nickname" validate:"required,max=10,alphanumspace,lowercase"`
}

type StartSoloGame struct {
	Nickname   string `json:"nickname" validate:"required,max=10,alphanumspace,lowercase"`
	Difficulty int    `json:"difficulty" validate:"oneof=0 1 2"`
}

type JoinGame struct {
	Nickname string `json:"nickname" validate:"required,max=10,alphanumspace,lowercase,nefield=Host"`
	Host     string `json:"host" validate:"required,max=10,alphanumspace,lowercase"`
}

type Joined struct {
	Nickname string `json:"nickname"`
}

type LeaveGame struct {
	Nickname string `json:"nickname" validate:"required,max=10,alphanumspace,lowercase"`
	Host     string `json:"host" validate:"required,max=10,alphanumspace,lowercase"`
}

type GameOver struct {
	Message string `json:"message"`
}

type ListOpenGames struct{}

type OpenGames struct {
	Hosts []string `json:"hosts"`
}

type PlaceDisk struct {
	Nickname string `json:"nickname" validate:"required,max=10,alphanumspace,lowercase"`
	Host     string `json:"host" validate:"required,max=10,alphanumspace,lowercase"`
	X        int    `json:"x" validate:"min=0,max=7"`
	Y        int    `json:"y" validate:"min=0,max=7"`
}

type UpdateBoard struct {
	Board   common.Board `json:"board"`
	Player  common.Disk  `json:"player"`
	X       int          `json:"x"`
	Y       int          `json:"y"`
	P1Score int          `json:"p1score"`
	P2Score int          `json:"p2score"`
}

type Error struct {
	Error string `json:"error"`
}

type Decorate struct {
	Decoration string `json:"decoration"`
}
