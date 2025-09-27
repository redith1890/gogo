package engine

import (
	"strings"
)

type EyeId int
type RegionId int
type ChainId int

// func make_grid(ysize int, xsize int) Grid {
// 	grid := make([][]Color, ysize)
// 	for i := 0; i < xsize; i++ {
// 		grid[i] = make([]Color, xsize)
// 	}
// 	return grid
// }

// func make_bool_grid(ysize int, xsize int) [][]bool {
// 	grid := make([][]bool, ysize)
// 	for i := 0; i < xsize; i++ {
// 		grid[i] = make([]bool, xsize)
// 	}
// 	return grid
// }

// func make_ids_grid(ysize, xsize int) [][]int {
// 	grid := make([][]int, ysize)
// 	for i := 0; i < ysize; i++ {
// 		grid[i] = make([]int, xsize)
// 		for j := 0; j < xsize; j++ {
// 			grid[i][j] = -1
// 		}
// 	}
// 	return grid
// }

func make_grid[T any](ysize, xsize int, init T) [][]T {
	grid := make([][]T, ysize)
	for i := 0; i < ysize; i++ {
		grid[i] = make([]T, xsize)
		for j := 0; j < xsize; j++ {
			grid[i][j] = init
		}
	}
	return grid
}


type LocScore struct {
	is_territory_for        Color
	belongs_to_seki_group   Color
	is_false_eye            bool
	is_unscorable_false_eye bool
	is_dame                 bool
	eye_value               int
}

type RegionInfo struct {
	region_id       int
	color           Color
	region_and_dame Set[[2]int]
	eyes            Set[EyeId]
}

type ChainInfo struct {
	chain_id       ChainId
	region_id      RegionId
	color          Color
	points         [][2]int
	neighbors      Set[ChainId]
	adjacents      Set[[2]int]
	liberties      Set[[2]int]
	is_marked_dead bool
}

func TerritoryScoring(stones Grid, marked_dead [][]bool, score_false_eyes bool) {
	xsize := len(stones)
	ysize := len(stones[0])

	connection_blocks := make([][]Color, ysize)
	for i := 0; i < xsize; i++ {
		connection_blocks[i] = make([]Color, xsize)
	}
	mark_connection_blocks(ysize, xsize, stones, marked_dead, connection_blocks)

	strict_reaches_black := make_grid[bool](ysize, xsize, false)
	strict_reaches_white := make_grid[bool](ysize, xsize, false)
	mark_reachability(ysize, xsize, stones, marked_dead, nil, strict_reaches_black, strict_reaches_white)

	reaches_black := make_grid[bool](ysize, xsize, false)
	reaches_white := make_grid[bool](ysize, xsize, false)
	mark_reachability(ysize, xsize, stones, marked_dead, connection_blocks, reaches_black, reaches_white)

	region_ids := make_grid[RegionId](ysize, xsize, -1)
	region_infos_by_id := map[RegionId]RegionInfo{}
	mark_regions(ysize, xsize, stones, marked_dead, connection_blocks, reaches_black, reaches_white, region_ids, region_infos_by_id)

	chain_ids := make_grid[ChainId](ysize, xsize, -1)
	chain_infos_by_id := map[ChainId]*ChainInfo{}
	mark_chains(ysize, xsize, stones, marked_dead, region_ids, chain_ids, chain_infos_by_id)

}

// In the original implementation chain_infos_by_data is a map[ChainId]ChainInfo, Itsn't map[ChainId]*ChainInfo
// but we need a pointer to make a copy and later assign it. For now I am using a pointer.
func mark_chains(ysize int, xsize int, stones Grid, marked_dead [][]bool, region_ids [][]RegionId, chain_ids [][]ChainId, chain_infos_by_id map[ChainId]*ChainInfo) {
	var fill_chain func(int, int, ChainId, Color, bool)
	fill_chain = func(y int, x int, with_id ChainId, color Color, is_marked_dead bool) {
		if !is_on_board(y, x, ysize, xsize) {
			return
		}
		if chain_ids[y][x] == with_id {
			return
		}
		if chain_ids[y][x] != -1 {
			other_id := chain_ids[y][x]
			chain_infos_by_id[other_id].neighbors.Add(with_id)
			chain_infos_by_id[with_id].neighbors.Add(other_id)
			chain_infos_by_id[with_id].adjacents.Add([2]int{y, x})
			if stones[y][x] == Empty {
				chain_infos_by_id[with_id].liberties.Add([2]int{y, x})
			}
			return
		}
		if stones[y][x] != color || marked_dead[y][x] != is_marked_dead {
			chain_infos_by_id[with_id].adjacents.Add([2]int{y, x})
			if stones[y][x] == Empty {
				chain_infos_by_id[with_id].liberties.Add([2]int{y, x})
			}
			return
		}
		chain_ids[y][x] = with_id
		chain_infos_by_id[with_id].points = append(chain_infos_by_id[with_id].points, [2]int{y, x})
		if chain_infos_by_id[with_id].region_id != region_ids[y][x] {
			chain_infos_by_id[with_id].region_id = -1
		}
		if !(color == Empty || region_ids[y][x] == chain_infos_by_id[with_id].region_id) {
		    panic("assertion failed: contiguous chain must match region")
		}
		fill_chain(y-1, x, with_id, color, is_marked_dead)
		fill_chain(y+1, x, with_id, color, is_marked_dead)
		fill_chain(y, x-1, with_id, color, is_marked_dead)
		fill_chain(y, x+1, with_id, color, is_marked_dead)
	}

	next_chain_id := 0
	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			if chain_ids[y][x] == -1 {
				chain_id := ChainId(next_chain_id)
				next_chain_id++
				color := stones[y][x]
				is_marked_dead := marked_dead[y][x]
				chain_infos_by_id[chain_id] = &ChainInfo{
					chain_id,
					region_ids[y][x],
					color,
					make([][2]int, 0),
					make(Set[ChainId]),
					make(Set[[2]int]),
					make(Set[[2]int]),
					is_marked_dead,
				}
				fill_chain(y, x, chain_id, color, is_marked_dead)
			}
		}
	}
}

func mark_regions(ysize int, xsize int, stones Grid, marked_dead [][]bool, connection_blocks Grid, reaches_black [][]bool, reaches_white [][]bool, region_ids [][]RegionId, region_infos_by_id map[RegionId]RegionInfo) {

	var fill_region func(int, int, RegionId, Color, [][]bool, [][]bool, [][]bool)
	fill_region = func(y int, x int, with_id RegionId, opp Color, reaches_pla [][]bool, reaches_opp [][]bool, visited [][]bool) {
		if !is_on_board(y, x, ysize, xsize) {
			return
		}
		if visited[y][x] {
			return
		}
		if region_ids[y][x] != -1 {
			return
		}
		if stones[y][x] == opp && !marked_dead[y][x] {
			return
		}

		visited[y][x] = true
		region_infos_by_id[with_id].region_and_dame.Add([2]int{y, x})
		if reaches_pla[y][x] && !reaches_opp[y][x] {
			region_ids[y][x] = with_id
		}
		if connection_blocks[y][x] == opp {
			return
		}

		fill_region(y-1, x, with_id, opp, reaches_pla, reaches_opp, visited)
		fill_region(y+1, x, with_id, opp, reaches_pla, reaches_opp, visited)
		fill_region(y, x-1, with_id, opp, reaches_pla, reaches_opp, visited)
		fill_region(y, x+1, with_id, opp, reaches_pla, reaches_opp, visited)
	}

	next_region_id := 0
	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			if reaches_black[y][x] && !reaches_white[y][x] && region_ids[y][x] == -1 {
				region_id := next_region_id
				next_region_id += 1
				region_infos_by_id[RegionId(region_id)] = RegionInfo{region_id, Black, make(Set[[2]int]), make(Set[EyeId])}
				visited := make_grid[bool](ysize, xsize, false)
				fill_region(y, x, RegionId(region_id), White, reaches_black, reaches_white, visited)
			}
			if reaches_white[y][x] && !reaches_black[y][x] && region_ids[y][x] == -1 {
				region_id := next_region_id
				next_region_id += 1
				region_infos_by_id[RegionId(region_id)] = RegionInfo{region_id, White, make(Set[[2]int]), make(Set[EyeId])}
				visited := make_grid[bool](ysize, xsize, false)
				fill_region(y, x, RegionId(region_id), Black, reaches_white, reaches_black, visited)
			}
		}
	}
}

func mark_connection_blocks(ysize int, xsize int, stones Grid, marked_dead [][]bool, connection_blocks Grid) {
	patterns := [][]string{
		{
			"pp",
			"@e",
			"pe",
		},
		{
			"ep?",
			"e@e",
			"ep?",
		},
		{
			"pee",
			"e@p",
			"pee",
		},
		{
			"?e?",
			"p@p",
			"xxx",
		},
		{
			"pp",
			"@e",
			"xx",
		},
		{
			"ep?",
			"e@e",
			"xxx",
		},
	}

	for _, pla := range []Color{White, Black} {
		opp := Opp(pla)
		orientations := []struct {
			pdydy, pdydx, pdxdy, pdxdx int
		}{
			{1, 0, 0, 1},
			{-1, 0, 0, 1},
			{1, 0, 0, -1},
			{-1, 0, 0, -1},
			{0, 1, 1, 0},
			{0, -1, 1, 0},
			{0, 1, -1, 0},
			{0, -1, -1, 0},
		}
		for _, orient := range orientations {
			for _, pattern := range patterns {
				pylen := len(pattern)
				pxlen := len(pattern[0])
				is_edge_pattern := strings.Contains(pattern[pylen-1], "x")

				if is_edge_pattern {
					pylen-- // We check the edge specially
				}

				y_range := make_range(0, ysize-1)
				x_range := make_range(0, xsize-1)

				if is_edge_pattern {
					if orient.pdydy == -1 {
						y_range = []int{pylen - 2}
					} else if orient.pdydy == 1 {
						y_range = []int{ysize - pylen}
					} else if orient.pdxdy == -1 {
						x_range = []int{pylen - 2}
					} else if orient.pdxdy == 1 {
						x_range = []int{xsize - pylen}
					}
				}

				for _, y := range y_range {
					for _, x := range x_range {
						get_target_yx := func(pdy int, pdx int) (int, int) {
							ty := y + orient.pdydy*pdy + orient.pdxdy*pdx
							tx := x + orient.pdydx*pdy + orient.pdxdx*pdx
							return ty, tx
						}
						ty, tx := get_target_yx(pylen-1, pxlen-1)
						if !is_on_board(ty, tx, ysize, xsize) || !is_on_board(y, x, ysize, xsize) {
							continue
						}

						var atloc Point
						mismatch := false

						for pdy := 0; pdy < pylen; pdy++ {
							for pdx := 0; pdx < pxlen && pdx < len(pattern[pdy]); pdx++ {
								c := pattern[pdy][pdx]
								ty, tx := get_target_yx(pdy, pdx)

								if !is_on_board(ty, tx, ysize, xsize) {
									mismatch = true
									break
								}

								switch c {
								case '?':
									continue
								case 'p':
									if !(stones[ty][tx] == pla && !marked_dead[ty][tx]) {
										mismatch = true
									}
								case 'e':
									stone := stones[ty][tx]
									if stone != Empty &&
										!(stone == pla && !marked_dead[ty][tx]) &&
										!(stone == opp && marked_dead[ty][tx]) {
										mismatch = true
									}
								case '@':
									if stones[ty][tx] != Empty {
										mismatch = true
									}
									atloc.Y, atloc.X = ty, tx
								case 'x':
									continue
								default:
									mismatch = true
								}

								if mismatch {
									break
								}
							}
							if mismatch {
								break
							}
						}
						if !mismatch && atloc.Y >= 0 && atloc.X >= 0 {
							connection_blocks[atloc.Y][atloc.X] = pla
						}
					}
				}
			}
		}
	}
}

func is_on_board(y int, x int, ysize int, xsize int) bool {
	return y >= 0 && x >= 0 && y < ysize && x < xsize
}

func make_range(start int, end int) []int {
	if start > end {
		return []int{}
	}
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

func AreaScoring(stones Grid, marked_dead [][]bool) Grid {
	xsize := len(stones)
	ysize := len(stones[0])

	strict_reaches_black := make_grid[bool](ysize, xsize, false)
	strict_reaches_white := make_grid[bool](ysize, xsize, false)
	scoring := make_grid[Color](ysize, xsize, Empty)

	mark_reachability(ysize, xsize, stones, marked_dead, nil, strict_reaches_black, strict_reaches_white)

	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			if strict_reaches_white[y][x] && !strict_reaches_black[y][x] {
				scoring[y][x] = White
			}
			if strict_reaches_black[y][x] && !strict_reaches_white[y][x] {
				scoring[y][x] = Black
			}
		}
	}
	return scoring
}

func mark_reachability(xsize int, ysize int, stones Grid, marked_dead [][]bool, connection_blocks Grid, reaches_black [][]bool, reaches_white [][]bool) {

	var fill_reach func(int, int, [][]bool, Color)
	
	fill_reach = func(y int, x int, reaches_pla [][]bool, pla Color) {
		if !is_on_board(y, x, ysize, xsize) {
			return
		}
		if reaches_pla[y][x] {
			return
		}
		if stones[y][x] == Opp(pla) && !marked_dead[y][x] {
			return
		}
		reaches_pla[y][x] = true
		if connection_blocks != nil && connection_blocks[y][x] == Opp(pla) {
			return
		}

		fill_reach(y-1, x, reaches_pla, pla)
		fill_reach(y+1, x, reaches_pla, pla)
		fill_reach(y, x-1, reaches_pla, pla)
		fill_reach(y, x+1, reaches_pla, pla)
	}

	for y := range ysize {
		for x := range xsize {
			if stones[y][x] == Black && !marked_dead[y][x] {
				fill_reach(y, x, reaches_black, Black)
			}
			if stones[y][x] == White && !marked_dead[y][x] {
				fill_reach(y, x, reaches_white, White)
			}
		}
	}

}
