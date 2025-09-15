package engine

import (
	. "fmt"
	"math/rand"
	// "go-online/ui"
)

const (
	Black = iota
	White
	Empty
	Random
)

type Piece struct {
	color int
	ko    bool
}

type Point struct {
	X int
	Y int
}

type Group struct {
	pos []Point
}

type Game struct {
	Grid  [][]int
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

	game := Game{Grid: new_grid,
		Turn:  start_turn,
		Score: [2]uint{0, 0},
	}
	return game

}

func opposite(color_turn int) int {
	if color_turn == White {
		return Black
	} else {
		return White
	}
}

func (game *Game) Move(pos Point) bool {
	Print(pos)
	Println("engine move: ", game.Turn)

	if !game.is_move_legal(pos) {
		Println("El movimiento %v no es legal", pos)
		return false
	}

	game.Grid[pos.X][pos.Y] = game.Turn
	game.Turn = opposite(game.Turn)

	return true
}

func (game *Game) is_move_legal(pos Point) bool {
	if game.Grid[pos.X][pos.Y] != Empty {
		Println(game.Grid[pos.X][pos.Y])
		return false
	}

	is_suicide := game.is_suicide(pos)
	is_eating := game.is_eating(pos)
	game.eat(pos)

	if is_suicide {
		if is_eating {
			Println("COME !TODO")
		} else {
			return false
		}
	}

	return true
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
				game.Grid[groups[i][point].X][groups[i][point].Y] = Empty
				count++
			}
			game.Score[game.Turn] += count
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

	size := len(game.Grid)

	up = pos.Y == 0 || game.Grid[pos.X][pos.Y-1] == opposite(game.Turn)
	down = pos.Y == size-1 || game.Grid[pos.X][pos.Y+1] == opposite(game.Turn)
	left = pos.X == 0 || game.Grid[pos.X-1][pos.Y] == opposite(game.Turn)
	right = pos.X == size-1 || game.Grid[pos.X+1][pos.Y] == opposite(game.Turn)

	if up && down && left && right {
		// _, liberties := game.SelectGroup(Point{pos.X - 1, pos.Y})
		// if len(liberties) == 1 {
		// 	return false
		// }
		// _, liberties = game.SelectGroup(Point{pos.X + 1, pos.Y})
		// if len(liberties) == 1 {
		// 	return false
		// }
		// _, liberties = game.SelectGroup(Point{pos.X, pos.Y - 1})
		// if len(liberties) == 1 {
		// 	return false
		// }
		// _, liberties = game.SelectGroup(Point{pos.X, pos.Y + 1})
		// if len(liberties) == 1 {
		// 	return false
		// }

		return true
	}

	if !up {
		_, liberties, _ := game.SelectGroup(Point{pos.X, pos.Y-1})
		if len(liberties) == 1 {
			return true
		}
	}
	if !down {
		_, liberties, _ := game.SelectGroup(Point{pos.X, pos.Y+1})
		if len(liberties) == 1 {
			return true
		}
	}
	if !left {
		_, liberties, _ := game.SelectGroup(Point{pos.X-1, pos.Y})
		if len(liberties) == 1 {
			return true
		}
	}
	if !right {
		_, liberties, _ := game.SelectGroup(Point{pos.X+1, pos.Y})
		if len(liberties) == 1 {
			return true
		}
	}
	return false
}

func (game *Game) MoveWithoutRules(pos Point, color int) bool {
	game.Grid[pos.X][pos.Y] = color
	return true
}

func IsOutOfRange(pos Point, size int) bool {
	if pos.X < 0 || pos.Y < 0 || pos.X >= size || pos.Y >= size {
		return true
	}
	return false
}

func (game *Game) SelectGroup(pos Point) ([]Point, []Point, int) {
	if IsOutOfRange(pos, len(game.Grid)) || game.Grid[pos.X][pos.Y] == Empty {
		return nil, nil, Empty
	}

	var points []Point
	var liberties []Point
	var visited = make(map[Point]bool)
	var color = game.Grid[pos.X][pos.Y]

	var traverse func(Point)
	traverse = func(p Point) {
		if visited[p] {
			return
		}
		if p.X < 0 || p.X >= len(game.Grid) || p.Y < 0 || p.Y >= len(game.Grid) {
			return
		}

		visited[p] = true

		if game.Grid[p.X][p.Y] == Empty {
			liberties = append(liberties, p)
			return
		}

		if game.Grid[p.X][p.Y] == game.Grid[pos.X][pos.Y] {
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

func Play() {
	game := NewGame(19, Black)

	// game.Move(Point{0,1})
	// game.Move(Point{10,1})
	// game.Move(Point{1,0})
	// game.Move(Point{0,0})

	game.MoveWithoutRules(Point{0, 0}, Black)
	game.MoveWithoutRules(Point{1, 1}, White)
	game.MoveWithoutRules(Point{0, 1}, White)
	game.MoveWithoutRules(Point{1, 0}, White)
	game.MoveWithoutRules(Point{3, 0}, White)

	group, liberties, _ := game.SelectGroup(Point{1, 1})
	Println(group, liberties)

	for _, row := range game.Grid {
		Println("%v", row)
	}

}
