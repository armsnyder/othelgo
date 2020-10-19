package common

func ApplyMove(board *Board, x int, y int, player int) bool {
	if board == nil {
		return false
	}

	if player != 1 && player != 2 {
		return false
	}

	// validate x and y are on the even board
	if x < 0 || x >= BoardSize || y < 0 || y >= BoardSize {
		return false
	}

	if WhoseTurn(*board) != player {
		return false
	}

	// verify cell is empty
	if board[x][y] != 0 {
		return false
	}
	// choose vectors
	updated := false

	for _, v := range [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}} {
		nextX := v[0] + x
		nextY := v[1] + y

		if nextX < 0 || nextX >= BoardSize || nextY < 0 || nextY >= BoardSize {
			continue
		}

		if board[nextX][nextY] == player%2+1 {
			// expand and aggregate vectors
			if expandVector(board, nextX, nextY, player, v) {
				board[x][y] = player
				updated = true
			}
		}
	}

	return updated
}

func expandVector(board *Board, x int, y int, player int, v [2]int) bool {
	// By the time expandVector is called, we have already chosen a vector from the position of the
	// placed disk that contains at least one of the opposing player's disks. Therefore, we need to
	// search along the vector for the next disk belonging to the current player.

	nextX := v[0] + x
	nextY := v[1] + y

	if nextX < 0 || nextX >= BoardSize || nextY < 0 || nextY >= BoardSize {
		return false
	}

	switch (*board)[nextX][nextY] {
	case 0:
		return false
	case player:
		(*board)[x][y] = player
		return true
	default:
		if expandVector(board, nextX, nextY, player, v) {
			(*board)[x][y] = player
			return true
		}

		return false
	}
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

func WhoseTurn(board Board) int {
	p1, p2 := KeepScore(board)
	return (p1+p2)%2 + 1
}

func GameOver(board Board) bool {
	p1, p2 := KeepScore(board)
	return p1+p2 == BoardSize*BoardSize
}
