package common_test

import (
	"testing"

	. "github.com/armsnyder/othelgo/pkg/common"
)

func buildTestBoard(p1, p2 [][2]int) (board Board) {
	for i, moves := range [][][2]int{p1, p2} {
		player := i + 1
		for _, move := range moves {
			x, y := move[0], move[1]
			board[x][y] = player
		}
	}
	return board
}

func TestApplyMove(t *testing.T) {
	type args struct {
		board  Board
		x      int
		y      int
		player int
	}
	tests := []struct {
		name        string
		args        args
		wantBoard   Board
		wantUpdated bool
	}{
		{
			name: "flip one short vector player 1",
			args: args{
				board: buildTestBoard(
					[][2]int{{1, 1}},
					[][2]int{{1, 2}},
				),
				x:      1,
				y:      3,
				player: 1,
			},
			wantBoard: buildTestBoard(
				[][2]int{{1, 1}, {1, 2}, {1, 3}},
				[][2]int{},
			),
			wantUpdated: true,
		},
		{
			name: "flip one short vector player 2",
			args: args{
				board: buildTestBoard(
					[][2]int{{1, 1}},
					[][2]int{{1, 2}},
				),
				x:      1,
				y:      0,
				player: 2,
			},
			wantBoard: buildTestBoard(
				[][2]int{},
				[][2]int{{1, 0}, {1, 1}, {1, 2}},
			),
			wantUpdated: true,
		},
		{
			name: "flip one long vector",
			args: args{
				board: buildTestBoard(
					[][2]int{{1, 1}, {1, 2}, {1, 3}},
					[][2]int{{1, 0}},
				),
				x:      1,
				y:      4,
				player: 2,
			},
			wantBoard: buildTestBoard(
				[][2]int{},
				[][2]int{{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}},
			),
			wantUpdated: true,
		},
		{
			name: "flip all vectors",
			args: args{
				board: buildTestBoard(
					[][2]int{{1, 1}, {1, 3}, {1, 5}, {3, 5}, {5, 5}, {5, 3}, {5, 1}, {3, 1}},
					[][2]int{{2, 2}, {2, 3}, {2, 4}, {3, 4}, {4, 4}, {4, 3}, {4, 2}, {3, 2}},
				),
				x:      3,
				y:      3,
				player: 1,
			},
			wantBoard: buildTestBoard(
				[][2]int{
					{1, 1}, {1, 3}, {1, 5}, {3, 5}, {5, 5}, {5, 3}, {5, 1}, {3, 1},
					{2, 2}, {2, 3}, {2, 4}, {3, 4}, {4, 4}, {4, 3}, {4, 2}, {3, 2},
					{3, 3},
				},
				[][2]int{},
			),
			wantUpdated: true,
		},
		{
			name: "illegal player number",
			args: args{
				board: buildTestBoard(
					[][2]int{{1, 1}},
					[][2]int{{1, 2}},
				),
				x:      1,
				y:      3,
				player: 3,
			},
			wantBoard: buildTestBoard(
				[][2]int{{1, 1}},
				[][2]int{{1, 2}},
			),
			wantUpdated: false,
		},
		{
			name: "illegal position off board",
			args: args{
				board: buildTestBoard(
					[][2]int{{1, 1}},
					[][2]int{{1, 2}},
				),
				x:      -4,
				y:      3,
				player: 1,
			},
			wantBoard: buildTestBoard(
				[][2]int{{1, 1}},
				[][2]int{{1, 2}},
			),
			wantUpdated: false,
		},
		{
			name: "illegal position no adjacent opponent",
			args: args{
				board: buildTestBoard(
					[][2]int{{3, 3}, {4, 4}},
					[][2]int{{3, 4}, {4, 3}},
				),
				x:      2,
				y:      2,
				player: 1,
			},
			wantBoard: buildTestBoard(
				[][2]int{{3, 3}, {4, 4}},
				[][2]int{{3, 4}, {4, 3}},
			),
			wantUpdated: false,
		},
		{
			name: "illegal position no bounding disk",
			args: args{
				board: buildTestBoard(
					[][2]int{},
					[][2]int{{1, 1}, {1, 2}},
				),
				x:      1,
				y:      3,
				player: 1,
			},
			wantBoard: buildTestBoard(
				[][2]int{},
				[][2]int{{1, 1}, {1, 2}},
			),
			wantUpdated: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBoard, gotUpdated := ApplyMove(tt.args.board, tt.args.x, tt.args.y, tt.args.player)
			if gotBoard != tt.wantBoard {
				t.Errorf("ApplyMove() got board = %v, want %v", gotBoard, tt.wantBoard)
			}
			if gotUpdated != tt.wantUpdated {
				t.Errorf("ApplyMove() got updated = %v, want %v", gotUpdated, tt.wantUpdated)
			}
		})
	}
}
