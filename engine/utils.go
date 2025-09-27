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
