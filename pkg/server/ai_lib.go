package server

import (
	"context"
	"log"
	"math"
)

// AIGameState represents the state of a game and implements game domain-specific logic.
type AIGameState interface {
	// Score evaluates the desirability of a state from the perspective of the AI player.
	Score() float64

	// AITurn returns true if the next move will be performed by the AI player.
	AITurn() bool

	// MoveCount returns the number of moves possible in the current state.
	MoveCount() int

	// Move performs the move at the given index and returns the next state after the move.
	Move(int) AIGameState
}

// minimaxWithIterativeDeepening invokes moveUsingMinimax multiple times using different search
// depths, up to a max depth of n. It blocks until the provided context's deadline is passed and
// then returns the result from the deepest moveUsingMinimax invocation. The result is a move index.
func minimaxWithIterativeDeepening(ctx context.Context, state AIGameState, n int) int {
	results := make(chan int)
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for i := 1; true; i++ {
			result, fullyExplored := moveUsingMinimax(state, i)
			results <- result
			if fullyExplored {
				close(results)
				return
			}

			select {
			case <-ctx2.Done():
				return
			default:
			}
		}
	}()

	result, _ := moveUsingMinimax(state, 0)
	var ok bool

	for i := 1; i < n; i++ {
		select {
		case result, ok = <-results:
			if !ok { // No more results.
				return result
			}
		case <-ctx.Done():
			return result
		}
	}

	return result
}

// moveUsingMinimax invokes minimax using the specified depth and then returns the best AI move.
func moveUsingMinimax(state AIGameState, n int) (int, bool) {
	log.Printf("Running moveUsingMinimax using n=%d", n)

	bestMove := 0
	bestScore := math.Inf(-1)
	fullyExplored := true

	for i := 0; i < state.MoveCount(); i++ {
		moveScore, moveFullyExplored := minimax(state.Move(i), n)

		if !moveFullyExplored {
			fullyExplored = false
		}

		if moveScore > bestScore {
			bestMove = i
			bestScore = moveScore
		}
	}

	log.Printf("moveUsingMinimax bestMove=%d, bestScore=%f, n=%d", bestMove, bestScore, n)

	return bestMove, fullyExplored
}

// minimax is the minimax adversarial search algorithm. It returns the score for an AIGameState
// after performing minimax up to the specified depth n, and a bool which is true if it fully
// fully explored the moves.
func minimax(state AIGameState, n int) (float64, bool) {
	if state.MoveCount() <= 0 {
		return state.Score(), true
	}

	if n <= 0 {
		return state.Score(), false
	}

	var (
		result     float64
		comparator func(float64) bool
	)

	if state.AITurn() {
		result = math.Inf(-1)
		comparator = func(v float64) bool { return v > result }
	} else {
		result = math.Inf(1)
		comparator = func(v float64) bool { return v < result }
	}

	fullyExplored := true

	for i := 0; i < state.MoveCount(); i++ {
		moveScore, moveFullyExplored := minimax(state.Move(i), n-1)

		if !moveFullyExplored {
			fullyExplored = false
		}

		if comparator(moveScore) {
			result = moveScore
		}
	}

	return result, fullyExplored
}
