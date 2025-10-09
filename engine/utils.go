package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

//Temporal

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

func (s Set[T]) Has(value T) bool {
	_, exists := s[value]
	return exists
}

func (s Set[T]) Remove(value T) {
	delete(s, value)
}

func assert(cond bool, msg string) {
	if !cond {
		// panic("Assertion failed: " + msg)
	}
}

func (s Set[T]) Update(new_set Set[T]) {
	for e := range new_set {
		s[e] = struct{}{}
	}
}

func (s Set[T]) List() []T {
	array := make([]T, 0, len(s))
	for pt := range s {
		array = append(array, pt)
	}
	return array
}

func (s Set[T]) Any(s2 Set[T]) bool {
	for pt := range s2 {
		if _, ok := s[pt]; ok {
			return true
		}
	}
	return false
}

func Exists[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func Copy[T comparable](s Set[T]) Set[T] {
	clone := make(Set[T], len(s))
	for k := range s {
		clone[k] = struct{}{}
	}
	return clone
}

func PrintGrid[T any](name string, grid [][]T) {
	fmt.Println("============================================================", name)
	for y := 0; y < len(grid); y++ {
		for x := 0; x < len(grid[y]); x++ {
			v := grid[y][x]
			switch b := any(v).(type) {
			case bool:
				if b {
					fmt.Print("1 ")
				} else {
					fmt.Print("0 ")
				}
			default:
				fmt.Printf("%v ", v)
			}
		}
		fmt.Println()
	}
}

func Log(var_name string, args ...interface{}) {
	if len(args) == 0 || len(args) > 2 {
		panic("log function only needs 1-2 args")
	}

	value := args[0]

	var file string
	if len(args) == 1 {
		file = "golang.log"
	} else {
		var ok bool
		file, ok = args[1].(string)
		if !ok {
			panic("second arg must be a string")
		}
	}

	_, filename, line, ok := runtime.Caller(1)
	if !ok {
		filename = "<unknown>"
		line = 0
	}

	filename = filepath.Base(filename)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("error opening log:", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "%s:%d | %s = %#v\n", filename, line, var_name, value)
}
