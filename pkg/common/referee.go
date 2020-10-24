package common

func ApplyMove(board Board, x int, y int, player int) (Board, bool) {
	updated := false
	vectors := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}

	for _, v := range vectors {
		if flipAlongVector(&board, x, y, player, v, 0) {
			updated = true
		}
	}

	return board, updated
}

func flipAlongVector(board *Board, x int, y int, player int, v [2]int, depth int) bool {
	if x < 0 || x >= BoardSize || y < 0 || y >= BoardSize {
		return false
	}

	piece := board[x][y]

	if depth > 0 && piece == 0 {
		return false
	}

	if depth > 1 && piece == player {
		return true
	}

	if flipAlongVector(board, x+v[0], y+v[1], player, v, depth+1) {
		board[x][y] = player
		return true
	}

	return false
}

func KeepScore(board Board) (p1 int, p2 int) {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			switch board[i][j] {
			case 1:
				p1++
			case 2:
				p2++
			}
		}
	}

	return p1, p2
}

func GameOver(board Board) bool {
	return !(HasMoves(board, 1) || HasMoves(board, 2))
}

func HasMoves(board Board, player int) bool {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if _, updated := ApplyMove(board, i, j, player); updated {
				return true
			}
		}
	}
	return false
}
