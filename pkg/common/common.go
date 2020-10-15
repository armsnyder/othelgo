package common

const (
	PlaceDiskAction   = "placeDisk"
	UpdateBoardAction = "updateBoard"
)

const BoardSize = 8

type Board [BoardSize][BoardSize]int

type (
	BaseMessage struct {
		Action string `json:"action"`
	}

	PlaceDiskMessage struct {
		Action string `json:"action"`
		Player int    `json:"player"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	}

	UpdateBoardMessage struct {
		Action string `json:"action"`
		Board  Board  `json:"board"`
	}
)
