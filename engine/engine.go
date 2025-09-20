package engine

import (
	. "fmt"
	"math/rand"
	// "gogo/ui"
)
type Color uint8
type Grid [][]int
// Empty has to be 0 or refactorize it, black and white have to be 1 and 2 but only because the implementation of opp()
const (
	Empty = iota
	Black
	White
	Random
)

var Pass = Point{-1, -1}

type Point struct {
	X int
	Y int
}

type Move struct {
	// This gonna be used to generate the sgf, with letters/bytes instead of numbers
	pos Point
	Color int
}

type Game struct {
	Board  [][]int
	Moves []Move
	Turn  int
	Eaten [2]uint
	Score [2]uint
}

func (p Point) Up() Point {
	if p.Y == 0 {
		return p
	}
	return Point{p.X, p.Y - 1}
}

func (p Point) Down() Point {
	return Point{p.X, p.Y + 1}
}

func (p Point) Left() Point {
	if p.X == 0 {
		return p
	}
	return Point{p.X - 1, p.Y}
}

func (p Point) Right() Point {
	return Point{p.X + 1, p.Y}
}

func NewGame(size int, start_turn int) Game {

	new_grid := make([][]int, size)
	for x := range new_grid {
		new_grid[x] = make([]int, size)
		for y := range new_grid[x] {
			new_grid[x][y] = Empty
		}
	}

	if start_turn == Random {
		n := rand.Intn(2)
		if n == 0 {
			start_turn = White
		} else {
			start_turn = Black
		}
	}

	game := Game{Board: new_grid,
		Turn:  start_turn,
		Score: [2]uint{0, 0},
	}
	return game

}

func opp(color int) int {
	return 3 - color
}

func (game *Game) add_move(pos Point) int {
	move := Move{pos, game.Turn}
	game.Moves = append(game.Moves, move)
	return len(game.Moves)
}

func (game *Game) Move(pos Point) bool {
	if pos != Pass {
		if !game.is_move_legal(pos) {
			Println("El movimiento %v no es legal", pos)
			return false
		}
		game.Board[pos.X][pos.Y] = game.Turn
	}
	game.add_move(pos)
	game.Turn = opp(game.Turn)

	return true
}

func (game *Game) get_started_turn() int {
	if len(game.Moves) <= 0 {
		return game.Turn
	}
	return game.Moves[0].Color
}

func Eq[T comparable](a, b [][]T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func (game *Game) Copy() Game {
	new_grid := make([][]int, len(game.Board))
	for i := range game.Board {
		new_grid[i] = append([]int(nil), game.Board[i]...)
	}

	new_moves := append([]Move(nil), game.Moves...)

	return Game {
		Board: new_grid,
		Moves: new_moves,
		Turn: game.Turn,
		Eaten: game.Eaten,
		Score: game.Score,
	}
}

const ko_repetition_len = 10
func (game *Game) is_ko(pos Point) bool {
	old_game := game.Copy()

	var repetition_len int
	if ko_repetition_len < len(game.Moves) {
		repetition_len = ko_repetition_len
	} else {
		repetition_len = len(game.Moves)
	}

	Println("======================================================")
	Println("======================================================")
	Println("======================================================")
	for i := 0; i < repetition_len; i++ {
		old_game.UndoLastMove()
		print(old_game.Board)
		Print(i)
		Println("======================================================")
		if Eq(game.Board, old_game.Board) {
			return true
		}
	}
	return false
}

func (game *Game) is_move_legal(pos Point) bool {


	if game.Board[pos.X][pos.Y] != Empty {
		Println(game.Board[pos.X][pos.Y])
		return false
	}

	// if game.is_ko(pos) {
	// 	return false
	// 	Println("ES OKOKOKOKOKOKO")
	// }

	is_suicide := game.is_suicide(pos)
	is_eating := game.is_eating(pos)
	game.eat(pos)

	if is_suicide && !is_eating {
		return false
	}

	return true
}

func clean_grid(grid [][]int) {
	for x := range grid {
		for y := range grid[x] {
			grid[x][y] = Empty
		}
	}
}

// Because this function calls Move this will lead to recursive exponencial uses of this function.
func (game *Game) UndoLastMove() {
	if len(game.Moves) > 0 {
		game.Moves = game.Moves[:len(game.Moves)-1]

		old_game := game.Copy()
		clean_grid(game.Board)
		first_turn := game.get_started_turn()
		game.Moves = []Move{}
		game.Score = [2]uint{}
		// game.Eaten = [2]uint{}
		game.Turn = int(first_turn)

		for _, move := range old_game.Moves {
			game.Move(move.pos)
		}
	}
}

func (game *Game) eat(pos Point) bool {
	var groups [4][]Point
	var liberties [4][]Point
	var to_eat [4]bool
	var color int


	groups[0], liberties[0], color = game.SelectGroup(Point{pos.X - 1, pos.Y})
	if len(liberties[0]) == 1 && color != game.Turn {
		to_eat[0] = true
	}
	groups[1], liberties[1], color = game.SelectGroup(Point{pos.X + 1, pos.Y})
	if len(liberties[1]) == 1 && color != game.Turn {
		to_eat[1] = true
	}
	groups[2], liberties[2], color = game.SelectGroup(Point{pos.X, pos.Y - 1})
	if len(liberties[2]) == 1 && color != game.Turn {
		to_eat[2] = true
	}
	groups[3], liberties[3], color = game.SelectGroup(Point{pos.X, pos.Y + 1})
	if len(liberties[3]) == 1 && color != game.Turn {
		to_eat[3] = true
	}

	for i := range to_eat {
		if to_eat[i] {
			var count uint = 0
			for point := range groups[i] {
				game.Board[groups[i][point].X][groups[i][point].Y] = Empty
				count++
			}
			game.Score[game.Turn-1] += count
		}
	}

	if to_eat[0] || to_eat[1] || to_eat[2] || to_eat[3] {
		return true
	}
	return false
}

func (game *Game) is_eating(pos Point) bool {
	var liberties [4][]Point
	var to_eat [4]bool
	var color int


	_, liberties[0], color = game.SelectGroup(Point{pos.X - 1, pos.Y})
	if len(liberties[0]) == 1 && color != game.Turn {
		to_eat[0] = true
	}
	_, liberties[1], color = game.SelectGroup(Point{pos.X + 1, pos.Y})
	if len(liberties[1]) == 1 && color != game.Turn {
		to_eat[1] = true
	}
	_, liberties[2], color = game.SelectGroup(Point{pos.X, pos.Y - 1})
	if len(liberties[2]) == 1 && color != game.Turn {
		to_eat[2] = true
	}
	_, liberties[3], color = game.SelectGroup(Point{pos.X, pos.Y + 1})
	if len(liberties[3]) == 1 && color != game.Turn {
		to_eat[3] = true
	}

	if to_eat[0] || to_eat[1] || to_eat[2] || to_eat[3] {
		return true
	}
	return false
}

func (game *Game) is_suicide(pos Point) bool {
	up, down, left, right := false, false, false, false

	size := len(game.Board)

	up = pos.Y == 0 || game.Board[pos.X][pos.Y-1] == opp(game.Turn)
	down = pos.Y == size-1 || game.Board[pos.X][pos.Y+1] == opp(game.Turn)
	left = pos.X == 0 || game.Board[pos.X-1][pos.Y] == opp(game.Turn)
	right = pos.X == size-1 || game.Board[pos.X+1][pos.Y] == opp(game.Turn)

	if up && down && left && right {
		return true
	}

	game.MoveWithoutRules(pos, game.Turn)
	_, liberties, _ := game.SelectGroup(pos)
	game.UndoLastMoveWithoutRules()
	if len(liberties) == 0 {
		return true
	}

	return false
}

func (game *Game) UndoLastMoveWithoutRules() {
	if len(game.Moves) > 0 {
		move := game.Moves[len(game.Moves)-1]
		game.Moves = game.Moves[:len(game.Moves)-1]
		game.Board[move.pos.X][move.pos.Y] = Empty
	}
}

// TODO: Move manage eat here
func (game *Game) MoveWithoutRules(pos Point, color int) {
	// Temporal.
	if pos != Pass {
		game.Board[pos.X][pos.Y] = game.Turn
	}
	game.add_move(pos)
	game.Board[pos.X][pos.Y] = color
}

func IsOutOfRange(pos Point, size int) bool {
	if pos.X < 0 || pos.Y < 0 || pos.X >= size || pos.Y >= size {
		return true
	}
	return false
}

func (game *Game) SelectGroup(pos Point) ([]Point, []Point, int) {
	if IsOutOfRange(pos, len(game.Board)) || game.Board[pos.X][pos.Y] == Empty {
		return nil, nil, Empty
	}

	var points []Point
	var liberties []Point
	var visited = make(map[Point]bool)
	var color = game.Board[pos.X][pos.Y]

	var traverse func(Point)
	traverse = func(p Point) {
		if visited[p] {
			return
		}
		if p.X < 0 || p.X >= len(game.Board) || p.Y < 0 || p.Y >= len(game.Board) {
			return
		}

		visited[p] = true

		if game.Board[p.X][p.Y] == Empty {
			liberties = append(liberties, p)
			return
		}

		if game.Board[p.X][p.Y] == game.Board[pos.X][pos.Y] {
			points = append(points, p)
			traverse(Point{p.X, p.Y - 1})
			traverse(Point{p.X, p.Y + 1})
			traverse(Point{p.X - 1, p.Y})
			traverse(Point{p.X + 1, p.Y})
		}
		return
	}
	traverse(pos)

	return points, liberties, color
}

func print(grid Grid) {
	for _, row := range grid {
		for _, cell := range row {
			switch cell {
			case 0:
				Print(". ")
			case 1:
				Print("B ")
			case 2:
				Print("W ")
			default:
				Print("? ")
			}
		}
		Println()
	}
}



func test_corner_area_scoring() {
	size := 9
	game := NewGame(size, Black)

	// game.MoveWithoutRules(Point{2, 2}, Black)
	// game.MoveWithoutRules(Point{2, 3}, Black)
	// game.MoveWithoutRules(Point{3, 2}, Black)
	// game.MoveWithoutRules(Point{4, 2}, Black)
	// game.MoveWithoutRules(Point{2, 4}, Black)
	// game.MoveWithoutRules(Point{1, 4}, Black)
	// game.MoveWithoutRules(Point{4, 1}, Black)
	// game.MoveWithoutRules(Point{0, 4}, Black)
	// game.MoveWithoutRules(Point{4, 0}, Black)

	// game.MoveWithoutRules(Point{2, 2}, Black)
	// game.MoveWithoutRules(Point{2, 0}, Black)
	// game.MoveWithoutRules(Point{2, 1}, Black)
	// game.MoveWithoutRules(Point{1, 2}, Black)
	// game.MoveWithoutRules(Point{0, 2}, Black)
	// game.MoveWithoutRules(Point{0, 0}, Black)
	// game.MoveWithoutRules(Point{1, 1}, Black)

	game.MoveWithoutRules(Point{0, 3},White)
	for i := range size {
		game.Board[i][3] = White
	}

	game.MoveWithoutRules(Point{0, 4}, Black)
	game.MoveWithoutRules(Point{1, 4}, Black)
	game.MoveWithoutRules(Point{1, 5}, Black)
	game.MoveWithoutRules(Point{1, 6}, Black)
	game.MoveWithoutRules(Point{1, 7}, Black)
	game.MoveWithoutRules(Point{0, 6}, Black)
	game.MoveWithoutRules(Point{1, 8}, Black)
	game.MoveWithoutRules(Point{0, 8}, Black)



	print(game.Board)
	Println("=====================================")

	marked_dead := make([][]bool, size)
	for i:= range marked_dead {
		marked_dead[i] = make([]bool, size)
	}

	result := AreaScoring(game.Board, marked_dead)

	print(result)





}

func Play() {
	test_corner_area_scoring()

}
