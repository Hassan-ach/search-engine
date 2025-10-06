package utils

import "fmt"

type Set[T comparable] struct {
	elements map[T]bool
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		elements: map[T]bool{},
	}
}

func (s *Set[T]) Add(t T) bool {
	_, ok := s.elements[t]
	if ok {
		return ok
	}
	s.elements[t] = true
	return true
}

func (s *Set[T]) BatchAdd(t ...T) {
	//
	for _, v := range t {
		s.elements[v] = true
	}
}

func (s *Set[T]) GetAll() []T {
	t := make([]T, len(s.elements))
	i := 0
	for v := range s.elements {
		t[i] = v
		i++
	}
	return t
}

func (s Set[T]) Contains(t T) bool {
	return s.elements[t]
}

func (s Set[T]) Print() {
	for url, ok := range s.elements {
		if ok {
			fmt.Println(url)
		}
	}
}
