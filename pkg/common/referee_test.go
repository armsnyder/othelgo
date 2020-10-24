package common_test

import (
	"testing"

	. "github.com/armsnyder/othelgo/pkg/common"
)

type move [2]int

func buildTestBoard(p1, p2 []move) (board Board) {
	for i, moves := range [][]move{p1, p2} {
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
		name           string
		args           args
		wantBoard      Board
		wantNotUpdated bool
	}{
		{
			name: "flip one short vector player 1",
			args: args{
				board: buildTestBoard(
					[]move{{1, 1}},
					[]move{{1, 2}},
				),
				x:      1,
				y:      3,
				player: 1,
			},
			wantBoard: buildTestBoard(
				[]move{{1, 1}, {1, 2}, {1, 3}},
				nil,
			),
		},
		{
			name: "flip one short vector player 2",
			args: args{
				board: buildTestBoard(
					[]move{{1, 1}},
					[]move{{1, 2}},
				),
				x:      1,
				y:      0,
				player: 2,
			},
			wantBoard: buildTestBoard(
				nil,
				[]move{{1, 0}, {1, 1}, {1, 2}},
			),
		},
		{
			name: "flip one long vector",
			args: args{
				board: buildTestBoard(
					[]move{{1, 1}, {1, 2}, {1, 3}},
					[]move{{1, 0}},
				),
				x:      1,
				y:      4,
				player: 2,
			},
			wantBoard: buildTestBoard(
				nil,
				[]move{{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}},
			),
		},
		{
			name: "flip all vectors",
			args: args{
				board: buildTestBoard(
					[]move{{1, 1}, {1, 3}, {1, 5}, {3, 5}, {5, 5}, {5, 3}, {5, 1}, {3, 1}},
					[]move{{2, 2}, {2, 3}, {2, 4}, {3, 4}, {4, 4}, {4, 3}, {4, 2}, {3, 2}},
				),
				x:      3,
				y:      3,
				player: 1,
			},
			wantBoard: buildTestBoard(
				[]move{
					{1, 1}, {1, 3}, {1, 5}, {3, 5}, {5, 5}, {5, 3}, {5, 1}, {3, 1},
					{2, 2}, {2, 3}, {2, 4}, {3, 4}, {4, 4}, {4, 3}, {4, 2}, {3, 2},
					{3, 3},
				},
				nil,
			),
		},
		{
			name: "illegal player number",
			args: args{
				board: buildTestBoard(
					[]move{{1, 1}},
					[]move{{1, 2}},
				),
				x:      1,
				y:      3,
				player: 3,
			},
			wantNotUpdated: true,
		},
		{
			name: "illegal move off board",
			args: args{
				board: buildTestBoard(
					[]move{{1, 1}},
					[]move{{1, 2}},
				),
				x:      -4,
				y:      3,
				player: 1,
			},
			wantNotUpdated: true,
		},
		{
			name: "illegal move no adjacent opponent",
			args: args{
				board: buildTestBoard(
					[]move{{3, 3}, {4, 4}},
					[]move{{3, 4}, {4, 3}},
				),
				x:      2,
				y:      2,
				player: 1,
			},
			wantNotUpdated: true,
		},
		{
			name: "illegal move no bounding disk",
			args: args{
				board: buildTestBoard(
					nil,
					[]move{{1, 1}, {1, 2}},
				),
				x:      1,
				y:      3,
				player: 1,
			},
			wantNotUpdated: true,
		},
		{
			name: "illegal position cell already occupied",
			args: args{
				board: buildTestBoard(
					nil,
					[]move{{1, 1}, {1, 2}},
				),
				x:      1,
				y:      2,
				player: 1,
			},
			wantNotUpdated: true,
		},
		{
			name: "illegal position board full",
			args: args{
				board: buildTestBoard(
					[]move{
						{4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}, {4, 5}, {4, 6}, {4, 7},
						{5, 0}, {5, 1}, {5, 2}, {5, 3}, {5, 4}, {5, 5}, {5, 6}, {5, 7},
						{6, 0}, {6, 1}, {6, 2}, {6, 3}, {6, 4}, {6, 5}, {6, 6}, {6, 7},
						{7, 0}, {7, 1}, {7, 2}, {7, 3}, {7, 4}, {7, 5}, {7, 6}, {7, 7},
					},
					[]move{
						{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7},
						{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7},
						{2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 7},
						{3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}, {3, 5}, {3, 6}, {3, 7},
					},
				),
				x:      1,
				y:      2,
				player: 1,
			},
			wantNotUpdated: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantNotUpdated {
				tt.wantBoard = tt.args.board
			}
			gotBoard, gotUpdated := ApplyMove(tt.args.board, tt.args.x, tt.args.y, tt.args.player)
			if gotBoard != tt.wantBoard {
				t.Errorf("ApplyMove() got board = %v, want %v", gotBoard, tt.wantBoard)
			}
			if gotUpdated == tt.wantNotUpdated {
				t.Errorf("ApplyMove() got updated = %v, want %v", gotUpdated, !tt.wantNotUpdated)
			}
		})
	}
}
