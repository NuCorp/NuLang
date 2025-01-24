package container

type Set[T comparable] map[T]struct{}

func (s Set[T]) Insert(v T) bool {
	if _, ok := s[v]; ok {
		return false
	}

	s[v] = struct{}{}

	return true
}
func (s Set[T]) Has(v T) bool {
	_, ok := s[v]

	return ok
}
