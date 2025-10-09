package engine

import (
	. "fmt"
	"strings"
	"testing"
)

func ParseBoard(s string) [][]Color {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	board := make([][]Color, len(lines))
	for y, line := range lines {
		board[y] = make([]Color, len(line))
		for x, c := range line {
			switch c {
			case '.':
				board[y][x] = Empty
			case 'B':
				board[y][x] = Black
			case 'W':
				board[y][x] = White
			}
		}
	}
	return board
}

func TestTerritory(t *testing.T) {
	const board = `
.BW......W.........
BBWWWWWWW..........
.BW...B............
.BW................
.B.................
.B.................
.B.................
..B................
..B................
..B................
..B................
BB.................
...................
...................
...................
...................
...................
...................
...................
`

	grid := ParseBoard(board)
	// t.Logf("territory len = x:%d y:%d", len(grid), len(grid[0]))
	// Grid(grid).Print()
	marked_dead := make_array2[bool](19, 19, false)
	scored := TerritoryScoring(grid, marked_dead, false)
	// t.Log(scored)

	for y := range scored {
		Print("\n")
		for x := range scored[y] {
			// t.Log(scored[y][x].IsTerritoryFor)
			Print(scored[y][x].IsTerritoryFor)
		}
	}

}
