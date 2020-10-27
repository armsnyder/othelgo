package server

import (
	"context"
	"errors"

	"github.com/armsnyder/othelgo/pkg/common"
)

type othelgoAIMove struct {
	x int
	y int
}

// doAIPlayerMove takes a turn as the AI player.
func doAIPlayerMove(ctx context.Context, game gameState) gameState {
	return newAI(othelgoScoreFn, othelgoPossibleMovesFn, othelgoApplyMoveFn)(ctx, game).(gameState)
}

func othelgoScoreFn(state interface{}) int {
	game := state.(gameState)
	_, p2 := common.KeepScore(game.board)
	return p2
}

func othelgoPossibleMovesFn(state interface{}) (moves []interface{}) {
	game := state.(gameState)

	for x := 0; x < common.BoardSize; x++ {
		for y := 0; y < common.BoardSize; y++ {
			if _, ok := common.ApplyMove(game.board, x, y, 2); ok {
				moves = append(moves, &othelgoAIMove{x: x, y: y})
			}
		}
	}

	return moves
}

func othelgoApplyMoveFn(state, move interface{}) interface{} {
	game := state.(gameState)

	if move != nil {
		aiMove := move.(*othelgoAIMove)
		board, updated := common.ApplyMove(game.board, aiMove.x, aiMove.y, 2)

		if !updated {
			panic(errors.New("board was not updated using AI move"))
		}

		game.board = board
	}

	game.player = game.player%2 + 1

	return game
}
