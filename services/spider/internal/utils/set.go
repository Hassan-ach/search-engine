package utils

type Set[T comparable] struct {
	Elements map[T]bool
}

func (s *Set[T]) Add(t T) bool {
	_, ok := s.Elements[t]
	if ok {
		return ok
	}
	s.Elements[t] = true
	return true
}

func (s Set[T]) Contains(t T) bool {
	_, ok := s.Elements[t]
	return ok
}
