package engine

// TODO:
// - Changes names of temporal variables (tpl)
//

import (
	. "fmt"
	"os"
	"strings"
)

type EyeId int
type RegionId int
type ChainId int
type MacroChainId int

func make_array2[T any](ysize, xsize int, init T) [][]T {
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
	IsTerritoryFor          Color
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

type MacroChainInfo struct {
	macrochain_id      ChainId
	region_id          RegionId
	color              Color
	points             [][2]int
	chains             Set[ChainId]
	eye_neighbors_from map[EyeId]Set[[2]int]
}

type EyeInfo struct {
	pla                       Color
	region_id                 RegionId
	eye_id                    EyeId
	potential_points          Set[[2]int]
	real_points               Set[[2]int]
	macrochain_neighbors_from map[MacroChainId]Set[[2]int]
	is_loose                  bool
	eye_value                 int
}

type EyePointInfo struct {
	adj_points                            [][2]int
	adj_eye_points                        [][2]int
	num_empty_adj_points                  int
	num_empty_adj_false_points            int
	num_empty_adj_eye_points              int
	num_opp_adj_false_points              int
	is_false_eye_poke                     bool
	num_moves_to_block                    int
	num_blockables_depending_on_this_spot int
}

func (region_info RegionInfo) Print() {
	Println("region_id: ", region_info.region_id)
	Println("color: ", region_info.color)
	Println("region_and_dame: ", region_info.region_and_dame)
	Println("eyes: ", region_info.eyes)
}

func TerritoryScoring(stones [][]Color, marked_dead [][]bool, score_false_eyes bool) [][]LocScore {
	xsize := len(stones)
	ysize := len(stones[0])
	os.Truncate("golang.log", 0)

	connection_blocks := make([][]Color, ysize)
	for i := 0; i < xsize; i++ {
		connection_blocks[i] = make([]Color, xsize)
	}
	mark_connection_blocks(ysize, xsize, stones, marked_dead, connection_blocks)
	PrintGrid("connection_blocks", connection_blocks)

	strict_reaches_black := make_array2[bool](ysize, xsize, false)
	strict_reaches_white := make_array2[bool](ysize, xsize, false)
	mark_reachability(ysize, xsize, stones, marked_dead, nil, strict_reaches_black, strict_reaches_white)
	PrintGrid("strict_reaches_black", strict_reaches_black)
	PrintGrid("strict_reaches_white", strict_reaches_white)

	reaches_black := make_array2[bool](ysize, xsize, false)
	reaches_white := make_array2[bool](ysize, xsize, false)
	mark_reachability(ysize, xsize, stones, marked_dead, connection_blocks, reaches_black, reaches_white)
	PrintGrid("reaches_black", reaches_black)
	PrintGrid("reaches_white", reaches_white)

	region_ids := make_array2[RegionId](ysize, xsize, -1)
	region_infos_by_id := map[RegionId]RegionInfo{}
	mark_regions(ysize, xsize, stones, marked_dead, connection_blocks, reaches_black, reaches_white, region_ids, region_infos_by_id)
	PrintGrid("region_ids", region_ids)
	for _, region_info := range region_infos_by_id {
		region_info.Print()
	}
	// PrintGrid("region_infos_by_id", region_infos_by_id)

	chain_ids := make_array2[ChainId](ysize, xsize, -1)
	chain_infos_by_id := map[ChainId]ChainInfo{}
	mark_chains(ysize, xsize, stones, marked_dead, region_ids, chain_ids, chain_infos_by_id)
	PrintGrid("chain_ids", chain_ids)
	// PrintGrid("chain_infos_by_id", chain_infos_by_id)

	macrochain_ids := make_array2[MacroChainId](ysize, xsize, -1)
	macrochain_infos_by_id := map[MacroChainId]MacroChainInfo{}
	mark_macrochains(ysize, xsize, stones, marked_dead, connection_blocks, region_ids, region_infos_by_id, chain_ids, chain_infos_by_id, macrochain_ids, macrochain_infos_by_id)
	PrintGrid("macrochain_ids", macrochain_ids)
	// PrintGrid("macrochain_infos_by_id", macrochain_infos_by_id)

	eye_ids := make_array2[EyeId](ysize, xsize, -1)
	eye_infos_by_id := map[EyeId]EyeInfo{}
	mark_potential_eyes(ysize, xsize, stones, marked_dead, strict_reaches_black, strict_reaches_white, region_ids, region_infos_by_id, macrochain_ids, macrochain_infos_by_id, eye_ids, eye_infos_by_id)
	PrintGrid("eye_ids", eye_ids)
	// PrintGrid("eye_infos_by_id", eye_infos_by_id)

	is_false_eye_point := make_array2[bool](ysize, xsize, false)
	mark_false_eye_points(ysize, xsize, region_ids, macrochain_ids, macrochain_infos_by_id, eye_infos_by_id, is_false_eye_point)
	PrintGrid("is_false_eye_point", is_false_eye_point)

	mark_eye_values(ysize, xsize, stones, marked_dead, region_ids, region_infos_by_id, chain_ids, chain_infos_by_id, is_false_eye_point, eye_ids, eye_infos_by_id)

	is_unscorable_false_eye_point := make_array2[bool](ysize, xsize, false)
	mark_false_eye_points(ysize, xsize, region_ids, macrochain_ids, macrochain_infos_by_id, eye_infos_by_id, is_unscorable_false_eye_point)
	PrintGrid("is_unscorable_false_eye_point", is_unscorable_false_eye_point)

	scoring := make([][]LocScore, ysize)
	for y := 0; y < ysize; y++ {
		scoring[y] = make([]LocScore, xsize)
		for x := 0; x < xsize; x++ {
			scoring[y][x] = make_locscore()
		}
	}

	mark_scoring(ysize, xsize, stones, marked_dead, score_false_eyes, strict_reaches_black, strict_reaches_white, region_ids, region_infos_by_id, chain_ids, chain_infos_by_id, is_false_eye_point, eye_ids, eye_infos_by_id, is_unscorable_false_eye_point, scoring)
	territory_for := make([][]Color, ysize)
	for y := 0; y < len(scoring); y++ {
		territory_for[y] = make([]Color, xsize)
		for x := 0; x < len(scoring[y]); x++ {
			territory_for[y][x] = scoring[y][x].IsTerritoryFor
		}
	}
	PrintGrid("TerritoryFor", territory_for)

	return scoring
}

func mark_scoring(ysize int, xsize int, stones [][]Color, marked_dead [][]bool, score_false_eyes bool, strict_reaches_black [][]bool, strict_reaches_white [][]bool, region_ids [][]RegionId, region_infos_by_id map[RegionId]RegionInfo, chain_ids [][]ChainId, chain_infos_by_id map[ChainId]ChainInfo, is_false_eye_point [][]bool, eye_ids [][]EyeId, eye_infos_by_id map[EyeId]EyeInfo, is_unscorable_false_eye_point [][]bool, scoring [][]LocScore) {
	extra_black_unscoreable_points := make(Set[[2]int], 0)
	extra_white_unscoreable_points := make(Set[[2]int], 0)
	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			if is_unscorable_false_eye_point[y][x] && stones[y][x] != Empty && marked_dead[y][x] {
				adjacents := [4][2]int{
					{y - 1, x},
					{y + 1, x},
					{y, x - 1},
					{y, x + 1},
				}
				if stones[y][x] == White {
					for _, point := range adjacents {
						extra_black_unscoreable_points.Add(point)
					}
				} else {
					for _, point := range adjacents {
						extra_white_unscoreable_points.Add(point)
					}
				}
			}
		}
	}

	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			s := scoring[y][x]
			region_id := region_ids[y][x]
			if region_id == -1 {
				s.is_dame = true
			} else {
				region_info := region_infos_by_id[region_id]
				color := region_info.color
				total_eyes := 0
				for eye_id := range region_info.eyes {
					total_eyes += eye_infos_by_id[eye_id].eye_value
				}
				if total_eyes <= 1 {
					s.belongs_to_seki_group = region_info.color
				}
				if is_false_eye_point[y][x] {
					s.is_false_eye = true
				}
				if is_unscorable_false_eye_point[y][x] {
					s.is_unscorable_false_eye = true
				}
				if (stones[y][x] == Empty || marked_dead[y][x]) && ((color == Black && extra_black_unscoreable_points.Has([2]int{y, x})) ||
					(color == White && extra_white_unscoreable_points.Has([2]int{y, x}))) {
					s.is_unscorable_false_eye = true
				}
				s.eye_value = 0
				if eye_ids[y][x] != -1 {
					s.eye_value = eye_infos_by_id[eye_ids[y][x]].eye_value
				}

				if (stones[y][x] != color || marked_dead[y][x]) &&
					s.belongs_to_seki_group == Empty &&
					(score_false_eyes || !s.is_unscorable_false_eye) &&
					chain_infos_by_id[chain_ids[y][x]].region_id == region_id &&
					!(color == White && strict_reaches_black[y][x]) &&
					!(color == Black && strict_reaches_white[y][x]) {
					s.IsTerritoryFor = color
				}
			}
		}
	}
}

func make_locscore() LocScore {
	return LocScore{
		IsTerritoryFor:          Empty,
		belongs_to_seki_group:   Empty,
		is_false_eye:            false,
		is_unscorable_false_eye: false,
		is_dame:                 false,
		eye_value:               0,
	}
}

func mark_eye_values(ysize int, xsize int, stones [][]Color, marked_dead [][]bool, region_ids [][]RegionId, region_infos_by_id map[RegionId]RegionInfo, chain_ids [][]ChainId, chain_infos_by_id map[ChainId]ChainInfo, is_false_eye_point [][]bool, eye_ids [][]EyeId, eye_infos_by_id map[EyeId]EyeInfo) {

	for _, eye_info := range eye_infos_by_id {
		pla := eye_info.pla
		opp := Opp(pla)

		info_by_point := make(map[[2]int]EyePointInfo)
		assert(len(eye_info.real_points) == 0, "eye_info shouldn't be filled yet")
		for tpl := range eye_info.potential_points {
			y, x := tpl[0], tpl[1]
			if !is_false_eye_point[y][x] {
				eye_info.real_points.Add([2]int{y, x})
				info := EyePointInfo{adj_points: make([][2]int, 0), adj_eye_points: make([][2]int, 0)}
				info_by_point[[2]int{y, x}] = info
			}
		}

		for tpl := range eye_info.real_points {
			y, x := tpl[0], tpl[1]
			info := info_by_point[[2]int{y, x}]
			adjacents := [4][2]int{
				{y - 1, x},
				{y + 1, x},
				{y, x - 1},
				{y, x + 1},
			}
			for _, adj := range adjacents {
				ay, ax := adj[0], adj[1]
				if !is_on_board(ay, ax, ysize, xsize) {
					continue
				}
				info.adj_points = append(info.adj_points, adj)
				if eye_info.real_points.Has(adj) {
					info.adj_eye_points = append(info.adj_eye_points, adj)
				}
			}
		}

		for tpl := range eye_info.real_points {
			y, x := tpl[0], tpl[1]
			info := info_by_point[[2]int{y, x}]
			for _, adj := range info.adj_points {
				ay, ax := adj[0], adj[1]
				if stones[ay][ax] == Empty {
					info.num_empty_adj_points++
				}
				if stones[ay][ax] == Empty && eye_info.real_points.Has(adj) {
					info.num_empty_adj_eye_points++
				}
				if stones[ay][ax] == Empty && is_false_eye_point[ay][ax] {
					info.num_empty_adj_false_points++
				}
				if stones[ay][ax] == opp && is_false_eye_point[ay][ax] {
					info.num_opp_adj_false_points++
				}
			}
			if info.num_opp_adj_false_points > 0 && stones[y][x] == opp {
				info.is_false_eye_poke = true
			}
			if info.num_empty_adj_false_points >= 2 && stones[y][x] == opp {
				info.is_false_eye_poke = true
			}
		}

		for tpl := range eye_info.real_points {
			y, x := tpl[0], tpl[1]
			info := info_by_point[[2]int{y, x}]
			info.num_moves_to_block = 0
			// info.num_moves_to_block_no_opps = 0

			// TODO: Optize this booleans
			for _, adj := range info.adj_points {
				ay, ax := adj[0], adj[1]
				block := 0
				if stones[ay][ax] == Empty && !Exists(eye_info.real_points, adj) {
					block = 1
				}
				if stones[ay][ax] == Empty && Exists(info_by_point, adj) && info_by_point[adj].num_opp_adj_false_points >= 1 {
					block = 1
				}
				if stones[ay][ax] == opp && Exists(info_by_point, adj) && info_by_point[adj].num_empty_adj_false_points >= 1 {
					block = 1
				}
				if stones[ay][ax] == opp && is_false_eye_point[ay][ax] {
					block = 1000
				}
				if stones[ay][ax] == opp && Exists(info_by_point, adj) && info_by_point[adj].is_false_eye_poke {
					block = 1000
				}
				info.num_moves_to_block += block
			}
			eye_value := 0

			for point := range eye_info.real_points {
				if pinfo, ok := info_by_point[point]; ok && pinfo.num_moves_to_block <= 1 {
					eye_value = 1
					break
				}
			}

			for point_to_delete := range eye_info.real_points {
				dy, dx := point_to_delete[0], point_to_delete[1]
				if !is_pseudolegal(ysize, xsize, stones, chain_ids, chain_infos_by_id, dy, dx, pla) {
					continue
				}
				set_point_to_delete := make(Set[[2]int], 0)
				set_point_to_delete.Add(point_to_delete)
				pieces := get_pieces(ysize, xsize, eye_info.real_points, set_point_to_delete)
				if len(pieces) < 2 {
					continue
				}

				should_bonus := info_by_point[point_to_delete].num_opp_adj_false_points == 1

				num_definite_eye_pieces := 0

				for _, piece := range pieces {
					zero_moves_to_block := false
					for point := range piece {
						if info_by_point[point].num_moves_to_block <= 0 {
							zero_moves_to_block = true
							break
						}
						if should_bonus && info_by_point[point].num_moves_to_block <= 1 {
							zero_moves_to_block = true
							break
						}
					}
					if zero_moves_to_block {
						num_definite_eye_pieces++
					}
				}
				eye_value = max(eye_value, num_definite_eye_pieces)
			}

			marked_dead_count := 0
			for point := range eye_info.real_points {
				y, x := point[0], point[1]
				if stones[y][x] == opp && marked_dead[y][x] {
					marked_dead_count++
				}
			}
			if marked_dead_count >= 5 {
				eye_value = max(eye_value, 1)
			}
			if marked_dead_count >= 8 {
				eye_value = max(eye_value, 2)
			}

			if eye_value < 2 {
				total := len(eye_info.real_points)
				count1 := 0
				count2 := 0
				count3 := 0
				for point := range eye_info.real_points {
					y, x := point[0], point[1]
					if info_by_point[point].num_moves_to_block >= 1 {
						count1++
					}
					if info_by_point[point].num_moves_to_block >= 2 {
						count2++
					}
					if stones[y][x] == opp && len(info_by_point[point].adj_eye_points) >= 2 {
						count3++
					}
				}
				if total-count1-count2-count3 >= 6 {
					eye_value = max(eye_value, 2)
				}
			}

			if eye_value < 2 {
				count1 := 0
				count2 := 0
				for point := range eye_info.real_points {
					y, x := point[0], point[1]
					if stones[y][x] == Empty && len(info_by_point[point].adj_eye_points) >= 4 {
						count1++
					}
					if stones[y][x] == Empty && len(info_by_point[point].adj_eye_points) >= 3 {
						count2++
					}
				}

				if count1+count2 >= 6 {
					eye_value = max(eye_value, 2)
				}
			}

			if eye_value < 2 {
				for point_to_delete := range eye_info.real_points {
					dy, dx := point_to_delete[0], point_to_delete[1]
					if stones[dy][dx] != Empty {
						continue
					}
					if is_on_border(dy, dx, ysize, xsize) {
						continue
					}
					if !is_pseudolegal(ysize, xsize, stones, chain_ids, chain_infos_by_id, dy, dx, pla) {
						continue
					}

					info1 := info_by_point[point_to_delete]
					if info1.num_moves_to_block > 1 || len(info1.adj_eye_points) < 3 {
						continue
					}

					for _, adjacent := range info1.adj_eye_points {
						info2 := info_by_point[adjacent]
						if len(info2.adj_eye_points) < 3 {
							continue
						}
						if info2.num_moves_to_block > 1 {
							continue
						}
						dy2, dx2 := adjacent[0], adjacent[1]
						if stones[dy2][dx2] != Empty && info2.num_empty_adj_eye_points <= 1 {
							continue
						}
						set_to_delete_and_adjacent := make(Set[[2]int], 0)
						set_to_delete_and_adjacent.Add(point_to_delete)
						set_to_delete_and_adjacent.Add(adjacent)
						pieces := get_pieces(ysize, xsize, eye_info.real_points, set_to_delete_and_adjacent)
						if len(pieces) < 2 {
							continue
						}

						num_definite_eye_pieces := 0
						num_double_definite_eye_pieces := 0
						for _, piece := range pieces {
							num_zero_moves_to_block := 0
							for point := range piece {
								if info_by_point[point].num_moves_to_block <= 0 {
									num_zero_moves_to_block++
									if num_zero_moves_to_block >= 2 {
										break
									}
								}
							}
							if num_zero_moves_to_block >= 1 {
								num_definite_eye_pieces++
							}
							if num_zero_moves_to_block >= 2 {
								num_double_definite_eye_pieces++
							}
						}
						if num_definite_eye_pieces >= 2 && num_double_definite_eye_pieces >= 1 && (stones[dy2][dx2] == Empty || num_double_definite_eye_pieces >= 2) {
							eye_value = max(eye_value, 2)
							break
						}
					}
					if eye_value >= 2 {
						break
					}
				}
			}

			if eye_value < 2 {
				dead_opps_in_eye := make(Set[[2]int], 0)
				unplayable_in_eye := make([][2]int, 0)

				for point := range eye_info.real_points {
					dy, dx := point[0], point[1]
					if stones[dy][dx] == opp && marked_dead[dy][dx] {
						dead_opps_in_eye.Add(point)
					} else if !is_pseudolegal(ysize, xsize, stones, chain_ids, chain_infos_by_id, dy, dx, pla) {
						unplayable_in_eye = append(unplayable_in_eye, point)
					}
				}
				if len(dead_opps_in_eye) > 0 {
					num_throwins := 0
					for point := range eye_info.potential_points {
						y, x := point[0], point[1]
						if stones[y][x] == opp && is_false_eye_point[y][x] {
							num_throwins++
						}
					}

					possible_omissions := make([][2]int, len(unplayable_in_eye))
					copy(possible_omissions, unplayable_in_eye)
					// TODO: Add None to omissions
					// possible_omissions = append(possible_omissions, nil)

					all_good_for_defender := true
					for _, omitted := range possible_omissions {
						remaining_shape := Copy(dead_opps_in_eye)
						for _, point := range unplayable_in_eye {
							if point != omitted {
								remaining_shape.Add(point)
							}
						}
						initial_piece_count := len(get_pieces(ysize, xsize, remaining_shape, make(Set[[2]int], 0)))
						num_bottlenecks := 0
						num_non_bottlenecks_high_degree := 0
						for point_to_delete := range remaining_shape {
							dy, dx := point_to_delete[0], point_to_delete[1]
							set_tmp := make(Set[[2]int], 0)
							set_tmp.Add(point_to_delete)
							if len(get_pieces(ysize, xsize, remaining_shape, set_tmp)) > initial_piece_count {
								num_bottlenecks++
							} else if count_adjacents_in(dy, dx, remaining_shape) >= 3 {
								num_non_bottlenecks_high_degree++
							}
						}
						bonus := 0
						if len(remaining_shape) >= 7 {
							bonus = 1
						}
						if initial_piece_count-num_throwins+(num_bottlenecks+num_non_bottlenecks_high_degree+bonus)/2 < 2 {
							all_good_for_defender = false
							break
						}
					}
					if all_good_for_defender {
						eye_value = 2
					}
				}
			}
			eye_value = min(eye_value, 2)
			eye_info.eye_value = eye_value
		}
	}
}

func count_adjacents_in(y int, x int, points Set[[2]int]) int {
	count := 0
	adjacents := [4][2]int{
		{y - 1, x},
		{y + 1, x},
		{y, x - 1},
		{y, x + 1},
	}
	for _, a := range adjacents {
		if points.Has(a) {
			count++
		}
	}
	return count
}

func is_on_border(y int, x int, ysize int, xsize int) bool {
	return y == 0 || x == 0 || y == ysize-1 || x == xsize-1
}

func get_pieces(ysize int, xsize int, points Set[[2]int], points_to_delete Set[[2]int]) []Set[[2]int] {
	used_points := make(Set[[2]int], 0)

	var floodfill func([2]int, Set[[2]int])
	floodfill = func(point [2]int, piece Set[[2]int]) {
		if used_points.Has(point) || points_to_delete.Has(point) {
			return
		}
		used_points.Add(point)
		piece.Add(point)

		y, x := point[0], point[1]
		adjacents := [4][2]int{
			{y - 1, x},
			{y + 1, x},
			{y, x - 1},
			{y, x + 1},
		}
		for _, point := range adjacents {
			if points.Has(point) {
				floodfill(point, points)
			}
		}
	}
	pieces := make([]Set[[2]int], 0)
	for point := range points {
		if !used_points.Has(point) {
			piece := make(Set[[2]int], 0)
			floodfill(point, piece)
			if len(piece) > 0 {
				pieces = append(pieces, piece)
			}
		}
	}
	return pieces

}

func is_pseudolegal(ysize int, xsize int, stones [][]Color, chain_ids [][]ChainId, chain_infos_by_id map[ChainId]ChainInfo, y int, x int, pla Color) bool {
	if stones[y][x] != Empty {
		return false
	}
	adjacents := [4][2]int{
		{y - 1, x},
		{y + 1, x},
		{y, x - 1},
		{y, x + 1},
	}
	opp := Opp(pla)
	for _, adj := range adjacents {
		ay, ax := adj[0], adj[1]
		if is_on_board(ay, ax, ysize, xsize) {
			if stones[ay][ax] != opp {
				return true
			}
			if len(chain_infos_by_id[chain_ids[ay][ax]].liberties) <= 1 {
				return true
			}
		}
	}
	return false
}

func mark_false_eye_points(ysize int, xsize int, region_ids [][]RegionId, macrochain_ids [][]MacroChainId, macrochain_infos_by_id map[MacroChainId]MacroChainInfo, eye_infos_by_id map[EyeId]EyeInfo, is_false_eye_point [][]bool) {
	for orig_eye_id, orig_eye_info := range eye_infos_by_id {
		for orig_macrochain_id, neighbors_from_eye_points := range orig_eye_info.macrochain_neighbors_from {
			for pt := range neighbors_from_eye_points {
				ey, ex := pt[0], pt[1]
				same_eye_adj_count := 0

				neighbors := func(ey int, ex int) [][2]int {
					return [][2]int{
						{ey - 1, ex},
						{ey + 1, ex},
						{ey, ex - 1},
						{ey, ex + 1},
					}
				}

				for _, p := range neighbors(ey, ex) {
					if _, ok := orig_eye_info.potential_points[p]; ok {
						same_eye_adj_count++
					}
				}

				if same_eye_adj_count > 1 {
					continue
				}
				reaching_sides := make(Set[[2]int])
				visited_macro := make(Set[MacroChainId])
				visited_other_eyes := make(Set[EyeId])
				visited_orig_eye_points := make(Set[[2]int])

				visited_orig_eye_points.Add([2]int{ey, ex})

				target_side_count := 0
				for _, pt2 := range neighbors(ey, ex) {
					y, x := pt2[0], pt2[1]

					if is_on_board(y, x, ysize, xsize) && region_ids[y][x] == orig_eye_info.region_id {
						target_side_count++
					}
				}

				var search func(MacroChainId) bool
				search = func(macrochain_id MacroChainId) bool {
					if visited_macro.Has(macrochain_id) {
						return false
					}
					visited_macro.Add(macrochain_id)

					macrochain_info := macrochain_infos_by_id[macrochain_id]
					for eye_id, neighbors_from_macro_points := range macrochain_info.eye_neighbors_from {
						if visited_other_eyes.Has(eye_id) {
							continue
						}
						if eye_id == orig_eye_id {
							eye_info := eye_infos_by_id[eye_id]

							for pt2 := range neighbors_from_macro_points {
								y, x := pt2[0], pt2[1]
								if is_adjacent(y, x, ey, ex) {
									reaching_sides.Add([2]int{y, x})
								}
							}
							if len(reaching_sides) >= target_side_count {
								return true
							}

							points_reached := find_recursively_adjacent_points(eye_info.potential_points, eye_info.macrochain_neighbors_from[macrochain_id], visited_orig_eye_points)
							if len(points_reached) == 0 {
								continue
							}
							visited_orig_eye_points.Update(points_reached)

							if eye_info.eye_value > 0 {
								for point := range points_reached {
									if eye_info.real_points.Has(point) {
										return true
									}
								}
							}

							for point := range points_reached {
								y, x := point[0], point[1]
								if is_adjacent(y, x, ey, ex) {
									reaching_sides.Add([2]int{y, x})
								}
							}

							if len(reaching_sides) >= target_side_count {
								return true
							}

							for next_macrochain_id, from_eye_points := range eye_info.macrochain_neighbors_from {
								if points_reached.Any(from_eye_points) {
									if search(next_macrochain_id) {
										return true
									}
								}
							}
						} else {
							visited_other_eyes.Add(eye_id)
							eye_info := eye_infos_by_id[eye_id]
							if eye_info.eye_value > 0 {
								return true
							}
							for next_macrochain_id := range eye_info.macrochain_neighbors_from {
								if search(next_macrochain_id) {
									return true
								}
							}
						}
					}
					return false
				}
				if search(orig_macrochain_id) {
					Println("Not a false eye TESTING MACRO")
				} else {
					is_false_eye_point[ey][ex] = true
				}
			}
		}
	}
}

func find_recursively_adjacent_points(within_set Set[[2]int], from_points Set[[2]int], excluding_points Set[[2]int]) Set[[2]int] {
	expanded := make(Set[[2]int])
	from_points_array := from_points.List()
	i := 0
	for i < len(from_points_array) {
		point := from_points_array[i]
		i++
		if excluding_points.Has(point) || expanded.Has(point) || !within_set.Has(point) {
			continue
		}
		expanded.Add(point)
		y, x := point[0], point[1]
		from_points_array = append(from_points_array, [2]int{y - 1, x})
		from_points_array = append(from_points_array, [2]int{y + 1, x})
		from_points_array = append(from_points_array, [2]int{y, x - 1})
		from_points_array = append(from_points_array, [2]int{y, x + 1})
	}
	return expanded
}

func is_adjacent(y1 int, x1 int, y2 int, x2 int) bool {
	return (y1 == y2 && (x1 == x2+1 || x1 == x2-1)) || (x1 == x2 && (y1 == y2+1 || y1 == y2-1))
}

func mark_potential_eyes(ysize int, xsize int, stones [][]Color, marked_dead [][]bool, strict_reaches_black [][]bool, strict_reaches_white [][]bool, region_ids [][]RegionId, region_infos_by_id map[RegionId]RegionInfo, macrochain_ids [][]MacroChainId, macrochain_infos_by_id map[MacroChainId]MacroChainInfo, eye_ids [][]EyeId, eye_infos_by_id map[EyeId]EyeInfo) {
	next_eye_id := 0

	visited := make_array2[bool](ysize, xsize, false)

	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			if visited[y][x] {
				continue
			}
			if eye_ids[y][x] != -1 {
				continue
			}
			if stones[y][x] != Empty && !marked_dead[y][x] {
				continue
			}
			region_id := region_ids[y][x]
			if region_id == -1 {
				continue
			}
			region_info := region_infos_by_id[region_id]
			pla := region_info.color
			is_loose := strict_reaches_white[y][x] && strict_reaches_black[y][x]

			eye_id := next_eye_id
			next_eye_id++

			potential_points := make(Set[[2]int], 0)
			macrochain_neighbors_from := make(map[MacroChainId]Set[[2]int])
			var acc_region func(int, int, int, int)
			acc_region = func(y int, x int, prevy int, prevx int) {
				if !is_on_board(y, x, ysize, xsize) {
					return
				}
				if visited[y][x] {
					return
				}
				if region_ids[y][x] != region_id {
					return
				}
				if macrochain_ids[y][x] != -1 {
					macrochain_id := macrochain_ids[y][x]
					if _, ok := macrochain_neighbors_from[macrochain_id]; !ok {
						macrochain_neighbors_from[macrochain_id] = make(Set[[2]int])
					}
					macrochain_neighbors_from[macrochain_id].Add([2]int{prevy, prevx})
					if _, ok := macrochain_infos_by_id[macrochain_id].eye_neighbors_from[EyeId(eye_id)]; !ok {
						macrochain_infos_by_id[macrochain_id].eye_neighbors_from[EyeId(eye_id)] = make(Set[[2]int])
					}
					macrochain_infos_by_id[macrochain_id].eye_neighbors_from[EyeId(eye_id)].Add([2]int{y, x})
				}
				if stones[y][x] != Empty && !marked_dead[y][x] {
					return
				}
				visited[y][x] = true
				eye_ids[y][x] = EyeId(eye_id)
				potential_points.Add([2]int{y, x})
				acc_region(y-1, x, y, x)
				acc_region(y+1, x, y, x)
				acc_region(y, x-1, y, x)
				acc_region(y, x+1, y, x)
			}
			assert(macrochain_ids[y][x] == -1, "macrochain_ids[y][x] have to be -1")
			acc_region(y, x, 10000, 10000)
			eye_infos_by_id[EyeId(eye_id)] = EyeInfo{pla, region_id, EyeId(eye_id), potential_points, make(Set[[2]int]), macrochain_neighbors_from, is_loose, 0}

			region_infos_by_id[region_id].eyes.Add(EyeId(eye_id))
		}
	}

}

func mark_macrochains(
	ysize int,
	xsize int,
	stones [][]Color,
	marked_dead [][]bool,
	connection_blocks [][]Color,
	region_ids [][]RegionId,
	region_infos_by_id map[RegionId]RegionInfo,
	chain_ids [][]ChainId,
	chain_infos_by_id map[ChainId]ChainInfo,
	macrochain_ids [][]MacroChainId,
	macrochain_infos_by_id map[MacroChainId]MacroChainInfo) {

	next_macrochain_id := 0

	for pla := range []Color{Black, White} {
		opp := Opp(Color(pla))

		chains_handled := make(Set[ChainId])
		visited := make_array2[bool](ysize, xsize, false)
		for chain_id, chain_info := range chain_infos_by_id {
			if chains_handled.Has(chain_id) {
				continue
			}
			if !(chain_info.color == Color(pla) && !chain_info.is_marked_dead) {
				continue
			}
			region_id := chain_info.region_id
			assert(region_id != -1, "region_id cannot be -1")

			macrochain_id := next_macrochain_id
			next_macrochain_id++

			points := make([][2]int, 0)
			chains := make(Set[ChainId])

			var walk_and_accumulate func(int, int)
			walk_and_accumulate = func(y int, x int) {
				if !is_on_board(y, x, ysize, xsize) {
					return
				}
				if visited[y][x] {
					return
				}
				visited[y][x] = true

				chain_id = chain_ids[y][x]
				chain_info = chain_infos_by_id[chain_id]
				should_recurse := false
				if stones[y][x] == Color(pla) && !marked_dead[y][x] {
					macrochain_ids[y][x] = MacroChainId(macrochain_id)
					points = append(points, [2]int{y, x})
					if !chains.Has(chain_id) {
						chains.Add(chain_id)
						chains_handled.Add(chain_id)
					}
					should_recurse = true
				} else if region_ids[y][x] == -1 && connection_blocks[y][x] != opp {
					should_recurse = true
				}
				if should_recurse {
					walk_and_accumulate(y-1, x)
					walk_and_accumulate(y+1, x)
					walk_and_accumulate(y, x-1)
					walk_and_accumulate(y, x+1)
				}
			}

			walk_and_accumulate(chain_info.points[0][0], chain_info.points[0][1])
			macrochain_infos_by_id[MacroChainId(macrochain_id)] = MacroChainInfo{ChainId(macrochain_id), region_id, Color(pla), points, chains, map[EyeId]Set[[2]int]{}}
		}
	}
}

// In the original implementation chain_infos_by_data is a map[ChainId]ChainInfo, Itsn't map[ChainId]*ChainInfo
// but we need a pointer to make a copy and later assign it. For now I am using a pointer.
func mark_chains(ysize int, xsize int, stones [][]Color, marked_dead [][]bool, region_ids [][]RegionId, chain_ids [][]ChainId, chain_infos_by_id map[ChainId]ChainInfo) {
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

		ci := chain_infos_by_id[with_id]
		ci.points = append(ci.points, [2]int{y, x})
		if ci.region_id != region_ids[y][x] {
			ci.region_id = -1
		}
		chain_infos_by_id[with_id] = ci

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
				chain_infos_by_id[chain_id] = ChainInfo{
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

func mark_regions(ysize int, xsize int, stones [][]Color, marked_dead [][]bool, connection_blocks [][]Color, reaches_black [][]bool, reaches_white [][]bool, region_ids [][]RegionId, region_infos_by_id map[RegionId]RegionInfo) {

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
				visited := make_array2[bool](ysize, xsize, false)
				fill_region(y, x, RegionId(region_id), White, reaches_black, reaches_white, visited)
			}
			if reaches_white[y][x] && !reaches_black[y][x] && region_ids[y][x] == -1 {
				region_id := next_region_id
				next_region_id += 1
				region_infos_by_id[RegionId(region_id)] = RegionInfo{region_id, White, make(Set[[2]int]), make(Set[EyeId])}
				visited := make_array2[bool](ysize, xsize, false)
				fill_region(y, x, RegionId(region_id), Black, reaches_white, reaches_black, visited)
			}
		}
	}
}

func mark_connection_blocks(ysize int, xsize int, stones [][]Color, marked_dead [][]bool, connection_blocks [][]Color) {
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

	for _, pla := range []Color{Black, White} {
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
						y_range = []int{len(pattern) - 2}
					} else if orient.pdydy == 1 {
						y_range = []int{ysize - (len(pattern) - 1)}
					} else if orient.pdxdy == -1 {
						x_range = []int{len(pattern) - 2}
					} else if orient.pdxdy == 1 {
						x_range = []int{xsize - (len(pattern) - 1)}
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

func AreaScoring(stones [][]Color, marked_dead [][]bool) [][]Color {
	xsize := len(stones)
	ysize := len(stones[0])

	strict_reaches_black := make_array2[bool](ysize, xsize, false)
	strict_reaches_white := make_array2[bool](ysize, xsize, false)
	scoring := make_array2[Color](ysize, xsize, Empty)

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

func mark_reachability(xsize int, ysize int, stones [][]Color, marked_dead [][]bool, connection_blocks [][]Color, reaches_black [][]bool, reaches_white [][]bool) {

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

	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			if stones[y][x] == Black && !marked_dead[y][x] {
				fill_reach(y, x, reaches_black, Black)
			}
			if stones[y][x] == White && !marked_dead[y][x] {
				fill_reach(y, x, reaches_white, White)
			}
		}
	}

}
