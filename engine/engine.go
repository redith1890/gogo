package engine


import(
	. "fmt"
	"math/rand"
)


const (
	Black = iota
	White
	Empty
	Random
)

type Piece struct {
	color int
	ko bool
}

type Point struct {
	X int
	Y int
}


type PointNode struct {
	point Point
	Up *PointNode
	Down *PointNode
	Left *PointNode
	Right *PointNode
}

type Group struct {
	pos []Point
}

// type Board struct {

// }

type Game struct {
	grid [][]int
	turn int
	eaten [2]uint
	score [2]uint
}

func (p Point) Up() Point {
	if (p.Y == 0) {
		return p
	}
	return Point{p.X, p.Y-1}
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
	for i := range new_grid {
		new_grid[i] = make([]int, size)
		for j := range new_grid[i] {
			new_grid[i][j] = Empty
		}
	}

	if(start_turn == Random) {
		n := rand.Intn(2)
		if(n == 0) {
			start_turn = White
		} else {
			start_turn = Black
		}
	}

	game := Game {
		grid: new_grid,
		turn: start_turn,
		score: [2]uint{0, 0},
	}
	return game

}

func opposite(color_turn int) int {
	if(color_turn == White) {
		return Black
	} else {
		return White
	}
}

func (game *Game) Move(pos Point) bool {
	Print(pos)
	Println(game.turn)

	if(!game.is_move_legal(pos)) {
		Println("El movimiento %v no es legal", pos)
		return false
	}

	game.grid[pos.Y][pos.X] = game.turn
	game.turn = opposite(game.turn)


	return true
}

func (game *Game) is_suicide(pos Point) bool {
	up, down, left, right := false, false, false, false

	size := len(game.grid)

	up    = pos.Y == 0      || game.grid[pos.Y][pos.X-1] == opposite(game.turn)
	down  = pos.Y == size-1 || game.grid[pos.Y][pos.X+1] == opposite(game.turn)
	left  = pos.X == 0      || game.grid[pos.Y-1][pos.X] == opposite(game.turn)
	right = pos.X == size-1 || game.grid[pos.Y+1][pos.X] == opposite(game.turn)



	if(up && down && left && right) {
		_, liberties := game.SelectGroup(Point{pos.Y, pos.X-1})
		if(len(liberties) == 1) {
			return false
		}
		_, liberties = game.SelectGroup(Point{pos.Y, pos.X+1})
		if(len(liberties) == 1) {
			return false
		}
		_, liberties = game.SelectGroup(Point{pos.Y-1, pos.X})
		if(len(liberties) == 1) {
			return false
		}
		_, liberties = game.SelectGroup(Point{pos.Y+1, pos.X})
		if(len(liberties) == 1) {
			return false
		}

		return true;
	}
	return false;
}

func (game *Game) is_move_legal(pos Point) bool {
	if(game.grid[pos.Y][pos.X] != Empty) {
		Println(game.grid[pos.Y][pos.X])
		return false
	}

	if(game.is_suicide(pos)) {
		Println("es suicidio")
		return false
	}

	return true
}

func (game *Game) MoveWithoutRules(pos Point, color int) bool {
	game.grid[pos.Y][pos.X] = color
	return true
}

func (game *Game) SelectGroup(pos Point) ([]Point, []Point) {
	if game.grid[pos.X][pos.Y] == Empty {
		return nil, nil
	}

	var points []Point
	var liberties []Point
	var visited = make(map[Point]bool)


	var traverse func(Point)
	traverse = func(p Point) {
		if visited[p] {
			return
		}
		if p.X < 0 || p.X >= len(game.grid) || p.Y < 0 || p.Y >= len(game.grid) {
			return
		}

		visited[p] = true

		if game.grid[p.Y][p.X] == Empty {
			liberties = append(liberties, p)
			return
		}

		if game.grid[p.Y][p.X] == game.grid[pos.Y][pos.X] {
			points = append(points, p)
			traverse(Point{p.X, p.Y - 1})
			traverse(Point{p.X, p.Y + 1})
			traverse(Point{p.X - 1, p.Y})
			traverse(Point{p.X + 1, p.Y})
		}
		return
	}
	traverse(pos)


	return points, liberties
}

func Play() {
	game := NewGame(19, Black)

	// game.Move(Point{0,1})
	// game.Move(Point{10,1})
	// game.Move(Point{1,0})
	// game.Move(Point{0,0})

	game.MoveWithoutRules(Point{0,0}, Black)
	game.MoveWithoutRules(Point{1,1}, White)
	game.MoveWithoutRules(Point{0,1}, White)
	game.MoveWithoutRules(Point{1,0}, White)
	game.MoveWithoutRules(Point{3,0}, White)

	group, liberties := game.SelectGroup(Point{1,1})
	Println(group, liberties)


	for _, row := range game.grid {
		Println("%v", row)
	}
}