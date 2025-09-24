package engine

func AreaScoring(stones [][]int, marked_dead [][]bool) [][]int {



	xsize := len(stones)
	ysize := len(stones[0])


	strict_reaches_black := make([][]bool, ysize)
	strict_reaches_white := make([][]bool, ysize)
	scoring := make([][]int, ysize) // Empty initialized
	for i := range strict_reaches_black {
		strict_reaches_black[i] = make([]bool, xsize)
		strict_reaches_white[i] = make([]bool, xsize)
		scoring[i] = make([]int, xsize)
	}
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

func is_on_board(y int, x int, ysize int, xsize int) bool {
	if y < 0 || y >= ysize || x < 0 || x >= xsize {
		return false
	}

	return true
}

func mark_reachability(
	xsize int,
	ysize int,
	stones [][]int,
	marked_dead [][]bool,
	connection_blocks [][]int,
	reaches_black [][]bool,
	reaches_white [][]bool) {

	var fill_reach func(int, int, [][]bool, int)
	fill_reach = func(y int, x int, reaches_pla [][]bool, pla int) {
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

		fill_reach(y-1,x,reaches_pla,pla)
		fill_reach(y+1,x,reaches_pla,pla)
		fill_reach(y,x-1,reaches_pla,pla)
		fill_reach(y,x+1,reaches_pla,pla)
	}
	for y := range ysize {
		for x := range xsize {
			if stones[y][x] == Black && !marked_dead[y][x]{
				fill_reach(y, x, reaches_black, Black)
			}
			if stones[y][x] == White && !marked_dead[y][x] {
				fill_reach(y, x, reaches_white, White)
			}
		}
	}

}