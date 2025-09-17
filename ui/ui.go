package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	."fmt"
	. "go-online/engine"
)

const board_size = 19
var VisualGrid [board_size][board_size]rl.Rectangle
var Grid [board_size][board_size]int
var mouse_position rl.Vector2


func get_grid_point(game Game, pos rl.Vector2) (Point, bool) {
	for x := range VisualGrid {
		for y := range VisualGrid[x]{
			if rl.CheckCollisionPointRec(mouse_position, VisualGrid[x][y]) {
				return Point{X: x, Y:y}, true
			}
		}
	}
	return Point{}, false
}

func get_grid_middle_pos(p Point) rl.Vector2 {
	pos := VisualGrid[p.X][p.Y]
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



	grid_debug_lines := false

	game := NewGame(board_size, Black)
	Println(game)

	for x := 0; x < board_size; x++ {
    for y := 0; y < board_size; y++ {
      rec_pos := rl.Rectangle{
          pos_point.X + (float32(x) * spacing),
          pos_point.Y + (float32(y) * spacing),
          pos_point.Width,
          pos_point.Height,
      }
      VisualGrid[x][y] = rec_pos
      Grid[x][y] = Empty
    }
	}

	for !rl.WindowShouldClose() {

		rl.BeginDrawing()
			mouse_position = rl.GetMousePosition()

			rl.ClearBackground(rl.Beige)

			if rl.IsKeyPressed(rl.KeySpace) {
				grid_debug_lines = !grid_debug_lines
			}

			if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
				pos, exists := get_grid_point(game, mouse_position)
				if(exists) {
					points, liberties, _ := game.SelectGroup(pos)
					for _, point := range points {
						Println(get_grid_middle_pos(point))
						rl.DrawCircleLinesV(get_grid_middle_pos(point), 1, rl.Yellow)
					}
					Println("points: ", points)
					Println("liberties: ", liberties)
				}
			}

			if rl.IsKeyPressed(rl.KeyP) {
				game.Move(Pass)

			}

			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				for row := range VisualGrid {
					for col := range VisualGrid[row]{
						if rl.CheckCollisionPointRec(mouse_position, VisualGrid[row][col]) {
							game.Move(Point{row, col})
						}
					}
				}
			}
			white_score := Sprintf("W %d", game.Score[White-1])
			black_score := Sprintf("B %d", game.Score[Black-1])
			rl.DrawText(white_score, 20, 5, 22, rl.Blue)
			rl.DrawText(black_score, 100, 5, 22, rl.Blue)

			// Quizas cambiar esto a rectangles
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

			for x := range VisualGrid {
				for y := range VisualGrid[x] {
					if grid_debug_lines {
						rl.DrawRectangleLinesEx(VisualGrid[x][y], 2, rl.Black)
					}
					switch game.Grid[x][y] {
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
			if rl.IsKeyPressed(rl.KeyEnter) {
				game.UndoLastMove()
			}
		rl.EndDrawing()
	}
}