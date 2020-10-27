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
			results <- moveUsingMinimax(state, i)

			select {
			case <-ctx2.Done():
				return
			default:
			}
		}
	}()

	result := moveUsingMinimax(state, 0)

	for i := 1; i < n; i++ {
		select {
		case result = <-results:
		case <-ctx.Done():
			return result
		}
	}

	return result
}

// moveUsingMinimax invokes minimax using the specified depth and then returns the best AI move.
func moveUsingMinimax(state AIGameState, n int) int {
	log.Printf("Running moveUsingMinimax using n=%d", n)

	bestMove := 0
	bestScore := math.Inf(-1)

	for i := 0; i < state.MoveCount(); i++ {
		moveScore := minimax(state.Move(i), n, math.Inf(-1), math.Inf(1))

		if moveScore > bestScore {
			bestMove = i
			bestScore = moveScore
		}
	}

	log.Printf("moveUsingMinimax bestMove=%d, bestScore=%f, n=%d", bestMove, bestScore, n)

	return bestMove
}

// minimax is the minimax adversarial search algorithm. It returns the score for an AIGameState
// after performing minimax up to the specified depth n.
func minimax(state AIGameState, n int, alpha, beta float64) float64 {
	if n <= 0 || state.MoveCount() <= 0 {
		return state.Score()
	}

	var (
		result          float64
		comparator      func(float64, float64) float64
		alphaBetaUpdate func(float64)
		alphaBetaBreak  func() bool
	)

	if state.AITurn() {
		result = math.Inf(-1)
		comparator = math.Max
		alphaBetaUpdate = func(v float64) { alpha = math.Max(alpha, v) }
		alphaBetaBreak = func() bool { return alpha >= beta }
	} else {
		result = math.Inf(1)
		comparator = math.Min
		alphaBetaUpdate = func(v float64) { beta = math.Min(beta, v) }
		alphaBetaBreak = func() bool { return beta <= alpha }
	}

	for i := 0; i < state.MoveCount(); i++ {
		moveScore := minimax(state.Move(i), n-1, alpha, beta)
		result = comparator(result, moveScore)
		alphaBetaUpdate(moveScore)
		if alphaBetaBreak() {
			break
		}
	}

	return result
}
