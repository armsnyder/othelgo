package common

import "strings"

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
