package utils

import "fmt"

type SetQueu[T comparable] struct {
	elements []T        // Queu order
	lookup   map[T]bool // set membership
}


// NewSetQueu initializes a new SetQueu
func NewSetQueu[T comparable]() *SetQueu[T] {
	return &SetQueu[T]{
		elements: []T{},
		lookup:   make(map[T]bool),
	}
}

// Push adds an element if not already present
func (s *SetQueu[T]) Push(vals ...T) {
	for _, val := range vals {
		if !s.lookup[val] {
			s.elements = append(s.elements, val)
			s.lookup[val] = true
		}
	}
}

// Pop removes the first element (FIFO) and returns it
func (s *SetQueu[T]) Pop() (T, bool) {
	var zero T
	if len(s.elements) == 0 {
		return zero, false
	}
	val := s.elements[0]
	s.elements = s.elements[1:]
	delete(s.lookup, val)
	return val, true
}

// Contains checks membership
func (s *SetQueu[T]) Contains(val T) bool {
	return s.lookup[val]
}

// Len returns the number of elements
func (s *SetQueu[T]) Len() int {
	return len(s.elements)
}

// Empty return if the SetQueu empty
func (s *SetQueu[T]) Empty() bool {
	return s.Len() == 0
}

func (s SetQueu[T]) Print() {
	for i, s := range s.elements {
		fmt.Printf("%d: %v\n", i+1, s)
	}
}
