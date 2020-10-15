package messages

const (
	PlaceDiskAction   = "placeDisk"
	UpdateBoardAction = "updateBoard"
)

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
		Action string  `json:"action"`
		Board  [][]int `json:"board"`
	}
)
