package engine

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
        panic("Assertion failed: " + msg)
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