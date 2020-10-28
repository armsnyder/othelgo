package server

import (
	"math"

	"github.com/armsnyder/othelgo/pkg/common"
)

// doAIPlayerMove takes a turn as the AI player.
func doAIPlayerMove(board common.Board, difficulty int) common.Board {
	aiState := &aiGameState{
		board:  board,
		player: 2,
	}

	var depth int
	switch difficulty {
	default:
		depth = 1
	case 1:
		depth = 4
	case 2:
		depth = 6
	}

	move := findMoveUsingMinimax(aiState, depth)
	return aiState.moves[move]
}

// aiGameState implements the othelgo domain-specific logic needed by the AI.
type aiGameState struct {
	board  common.Board
	player common.Disk
	moves  []common.Board
}

func (a *aiGameState) Score() float64 {
	p1, p2 := common.KeepScore(a.board)

	if common.GameOver(a.board) {
		switch {
		case p2 > p1:
			return math.Inf(1)
		case p1 < p2:
			return math.Inf(-1)
		default:
			return 0
		}
	}

	trueScoreDelta := float64(p2 - p1)
	scoreModifier := a.scoreModifier(2) - a.scoreModifier(1)

	// Modifier strength decreases as the board fills up.
	scoreModifier *= a.percentFull()

	return trueScoreDelta + scoreModifier
}

func (a *aiGameState) scoreModifier(player common.Disk) (score float64) {
	endIndex := common.BoardSize - 1

	// Edges are valuable.
	edgeScore := 0.5
	for i := 1; i < endIndex; i++ {
		if a.board[i][0] == player {
			score += edgeScore
		}
		if a.board[0][i] == player {
			score += edgeScore
		}
		if a.board[i][endIndex] == player {
			score += edgeScore
		}
		if a.board[endIndex][i] == player {
			score += edgeScore
		}
	}

	// Corners are highly valuable.
	cornerScore := float64(2)
	if a.board[0][0] == player {
		score += cornerScore
	}
	if a.board[0][endIndex] == player {
		score += cornerScore
	}
	if a.board[endIndex][0] == player {
		score += cornerScore
	}
	if a.board[endIndex][endIndex] == player {
		score += cornerScore
	}

	return score
}

func (a *aiGameState) percentFull() float64 {
	freeCells := 0
	for x := 0; x < common.BoardSize; x++ {
		for y := 0; y < common.BoardSize; y++ {
			if a.board[x][y] == 0 {
				freeCells++
			}
		}
	}
	return float64(freeCells) / common.BoardSize / common.BoardSize
}

func (a *aiGameState) AITurn() bool {
	return a.player == 2
}

func (a *aiGameState) MoveCount() int {
	if a.moves == nil {
		a.moves = []common.Board{}
		for x := 0; x < common.BoardSize; x++ {
			for y := 0; y < common.BoardSize; y++ {
				if board, updated := common.ApplyMove(a.board, x, y, a.player); updated {
					a.moves = append(a.moves, board)
				}
			}
		}
	}

	return len(a.moves)
}

func (a *aiGameState) Move(i int) AIGameState {
	a.MoveCount() // Lazy initialize moves

	nextState := &aiGameState{
		board:  a.moves[i],
		player: a.player,
	}

	if common.HasMoves(a.moves[i], a.player%2+1) {
		nextState.player = a.player%2 + 1
	}

	return nextState
}
