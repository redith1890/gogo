package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	."fmt"
	. "gogo/engine"
)

const board_size = 19
var VisualBoard [board_size][board_size]rl.Rectangle
var mouse_position rl.Vector2


func get_grid_point(game Game, pos rl.Vector2) (Point, bool) {
	for x := range VisualBoard {
		for y := range VisualBoard[x]{
			if rl.CheckCollisionPointRec(mouse_position, VisualBoard[x][y]) {
				return Point{X: x, Y: y}, true
			}
		}
	}
	return Point{}, false
}

func get_grid_middle_pos(p Point) rl.Vector2 {
	pos := VisualBoard[p.X][p.Y]
	return rl.Vector2{pos.X + pos.Width/2, pos.Y + pos.Height/2}
}

func Draw() {
	rl.InitWindow(900, 900, "GoGo")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)
	const board_pixels = 800
	// const spacing_grid_lines = board_pixels / 18
	const spacing = board_pixels / (board_size - 1)
	const start = 50
	const end   = start + spacing*(board_size-1)
	pos_point := rl.Rectangle{
		(50-spacing) + spacing/2,
		(50-spacing) + spacing/2,
		spacing,
		spacing,
	}

	var fps int32



	grid_debug_lines := false
	scoring := false

	game := NewGame(board_size, Black)

	for x := 0; x < board_size; x++ {
    for y := 0; y < board_size; y++ {
      rec_pos := rl.Rectangle{
          pos_point.X + (float32(x) * spacing),
          pos_point.Y + (float32(y) * spacing),
          pos_point.Width,
          pos_point.Height,
      }
      VisualBoard[x][y] = rec_pos
    }
	}

	for !rl.WindowShouldClose() {

		rl.BeginDrawing()
			fps = rl.GetFPS()
			rl.DrawText(Sprintf("%d", fps), 830, 10, 22, rl.Green)
			mouse_position = rl.GetMousePosition()

			rl.ClearBackground(rl.Beige)

			if rl.IsKeyPressed(rl.KeyD) {
				grid_debug_lines = !grid_debug_lines
			}
			if rl.IsKeyPressed(rl.KeyC) {
				scoring = !scoring
			}
			if rl.IsKeyPressed(rl.KeyP) {
				game.Board.Print()
				for i := range game.Moves {
					game.Moves[i].Print()
				}
			}

			if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
				pos, exists := get_grid_point(game, mouse_position)
				if(exists) {
					points, liberties, _ := game.SelectGroup(pos)
					for _, point := range points {
						rl.DrawCircleLinesV(get_grid_middle_pos(point), 1, rl.Yellow)
					}
					Println("points: ", points)
					Println("liberties: ", liberties)
				}
			}

			if rl.IsKeyPressed(rl.KeySpace) {
				game.Move(Pass)
			}

			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				for x := range VisualBoard {
					for y := range VisualBoard[x]{
						if rl.CheckCollisionPointRec(mouse_position, VisualBoard[x][y]) {
							game.Move(Point{x, y})
						}
					}
				}
			}
			white_score := Sprintf("W %d", game.Score[White-1])
			black_score := Sprintf("B %d", game.Score[Black-1])
			rl.DrawText(white_score, 20, 5, 22, rl.Blue)
			rl.DrawText(black_score, 100, 5, 22, rl.Blue)

			// Maybe using rectangles it's better
			for i := 0; i < board_size; i++ {
				// Horizontal
				rl.DrawLineEx(
					rl.Vector2{start, float32(start) + float32(i) * spacing},
					rl.Vector2{end, float32(start) + float32(i) * spacing},
					2, rl.Red)
				// Vertical
				rl.DrawLineEx(
					rl.Vector2{float32(start) + float32(i) * spacing, start},
					rl.Vector2{float32(start) + float32(i) * spacing, end},
					2, rl.Red)
			}

			for x := range VisualBoard {
				for y := range VisualBoard[x] {
					if grid_debug_lines {
						rl.DrawRectangleLinesEx(VisualBoard[x][y], 2, rl.Black)
					}
					switch game.Board[y][x] {
						case Black:
							middle_point := get_grid_middle_pos(Point{x, y})
							rl.DrawCircleV(middle_point, spacing/2, rl.Black)
						case White:
							middle_point := get_grid_middle_pos(Point{x, y})
							rl.DrawCircleV(middle_point, spacing/2, rl.White)
						default:;
					}
				}
			}
			if scoring {
				marked_dead := make([][]bool, board_size)
				for i := range len(marked_dead) {
					marked_dead[i] = make([]bool, board_size)
				}
				scored := AreaScoring(game.Board, marked_dead)
				for y := range scored {
					for x := range scored[y] {
						if scored[y][x] == White {
							rl.DrawRectangleRec(VisualBoard[x][y], rl.LightGray)
						}
						if scored[y][x] == Black {
							rl.DrawRectangleRec(VisualBoard[x][y], rl.DarkGray)
						}
					}
				}
			}
			if rl.IsKeyPressed(rl.KeyU) {
				game.UndoLastMove()
				game.Turn = Opp(game.Turn)
			}
		rl.EndDrawing()
	}
}