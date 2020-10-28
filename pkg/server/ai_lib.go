package server

import (
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

// findMoveUsingMinimax invokes minimax using the specified depth and then returns the best AI move.
func findMoveUsingMinimax(state AIGameState, depth int) int {
	log.Printf("Running findMoveUsingMinimax using depth=%d", depth)

	bestMove := 0
	bestScore := math.Inf(-1)

	for i := 0; i < state.MoveCount(); i++ {
		moveScore := minimax(state.Move(i), depth, math.Inf(-1), math.Inf(1))

		if moveScore > bestScore {
			bestMove = i
			bestScore = moveScore
		}
	}

	log.Printf("findMoveUsingMinimax bestMove=%d, bestScore=%f, depth=%d", bestMove, bestScore, depth)

	return bestMove
}

// minimax is the minimax adversarial search algorithm. It returns the score for an AIGameState
// after performing minimax up to the specified depth n.
func minimax(state AIGameState, depth int, alpha, beta float64) float64 {
	if depth <= 0 || state.MoveCount() <= 0 {
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
		moveScore := minimax(state.Move(i), depth-1, alpha, beta)
		result = comparator(result, moveScore)
		alphaBetaUpdate(moveScore)
		if alphaBetaBreak() {
			break
		}
	}

	return result
}
