package server

import (
	"context"
)

// ai is a general function for modifying a game state using an AI player. The context can be used
// to set a timeout on the AI player's move. After the timeout expires the AI player will make the
// best move it calculated within the allotted timeframe and return.
type ai func(ctx context.Context, state interface{}) (stateOut interface{})

// scoreFn is a function that evaluates the desirability of a game state from the perspective of
// an AI player.
type scoreFn func(state interface{}) int

// possibleMovesFn is a function that lists the possible moves the AI player can make given a game
// state.
type possibleMovesFn func(state interface{}) (moves []interface{})

// applyMoveFn is a function that applies one of the moves returned from possibleMovesFn to the game
// state. The move argument may be nil if no move is possible.
type applyMoveFn func(state interface{}, move interface{}) (stateOut interface{})

// newAI is a general function for building a new AI player. The arguments are functions that
// implement the game domain-specific logic.
func newAI(scoreFn scoreFn, possibleMovesFn possibleMovesFn, applyMoveFn applyMoveFn) ai {
	return func(ctx context.Context, state interface{}) (move interface{}) {
		bestState := applyMoveFn(state, nil)
		bestStateScore := scoreFn(bestState)

		for _, move := range possibleMovesFn(state) {
			thisState := applyMoveFn(state, move)
			thisStateScore := scoreFn(thisState)

			if thisStateScore > bestStateScore {
				bestState = thisState
				bestStateScore = thisStateScore
			}
		}

		return bestState
	}
}
