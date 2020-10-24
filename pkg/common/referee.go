package common

func ApplyMove(board Board, x int, y int, player int) (Board, bool) {
	updated := false
	vectors := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}

	if isInBounds(x, y) && board[x][y] != 0 {
		return board, false
	}

	for _, v := range vectors {
		if flipAlongVector(&board, x+v[0], y+v[1], player, v, 0) {
			updated = true
			board[x][y] = player
		}
	}

	return board, updated
}

func flipAlongVector(board *Board, x int, y int, player int, v [2]int, depth int) bool {
	if !isInBounds(x, y) {
		return false
	}

	disk := board[x][y]

	switch disk {
	case 0:
		return false
	case player:
		return depth > 0
	}

	if flipAlongVector(board, x+v[0], y+v[1], player, v, depth+1) {
		board[x][y] = player
		return true
	}

	return false
}

func isInBounds(x int, y int) bool {
	return x >= 0 && x < BoardSize && y >= 0 && y < BoardSize
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
	if (board == Board{}) {
		return false
	}
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
