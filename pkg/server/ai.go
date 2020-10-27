package server

import (
	"context"
	"fmt"
	"time"

	"github.com/armsnyder/othelgo/pkg/common"
)

// doAIPlayerMove takes a turn as the AI player.
func doAIPlayerMove(ctx context.Context, board common.Board, player common.Disk, difficulty int) common.Board {
	aiState := &aiGameState{
		board:  board,
		player: player,
	}

	ctx2, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	var depth int
	switch difficulty {
	default:
		depth = 1
	case 1:
		depth = 5
	case 2:
		depth = common.BoardSize * common.BoardSize // No limit
	}

	move := minimaxWithIterativeDeepening(ctx2, aiState, depth)
	return aiState.moves[move]
}

type aiGameState struct {
	board  common.Board
	player common.Disk
	moves  []common.Board
}

func (a *aiGameState) Score() float64 {
	p1, p2 := common.KeepScore(a.board)
	switch a.player {
	case 1:
		return float64(p1)
	case 2:
		return float64(p2)
	default:
		panic(fmt.Errorf("illegal player %v", a.player))
	}
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
