package server

import (
	"fmt"
	"math"
	"testing"
)

func BenchmarkMiniMax(b *testing.B) {
	for _, depth := range []int{1, 2, 4} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			// Enable memory allocation stats for this benchmark test.
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var state aiGameState

				// New board.
				state.board[3][3] = 1
				state.board[4][4] = 1
				state.board[3][4] = 2
				state.board[4][3] = 2

				// Player 1 made the first move.
				state.board[2][4] = 1

				// Now it's player 2's turn (the AI player).
				state.player = 2

				// Do the thing being benchmarked.
				minimax(&state, depth, math.Inf(-1), math.Inf(1))
			}
		})
	}
}
