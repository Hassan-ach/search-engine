package utils

import "fmt"

type Stack[T any] struct {
	Elements []T
}

func (s *Stack[T]) Push(t ...T) {
	s.Elements = append(s.Elements, t...)
}

func (s *Stack[T]) Pop() (T, bool) {
	var t T
	if len(s.Elements) == 0 {
		return t, false
	}
	t = s.Elements[len(s.Elements)-1]
	s.Elements = s.Elements[:len(s.Elements)-1]
	return t, true
}

func (s Stack[T]) Peek() (T, bool) {
	var t T
	if len(s.Elements) == 0 {
		return t, false
	}
	return s.Elements[len(s.Elements)-1], true
}

func (s Stack[T]) Size() int {
	return len(s.Elements)
}

func (s Stack[T]) Empty() bool {
	return s.Size() == 0
}

func (s Stack[T]) Print() {
	for _, e := range s.Elements {
		fmt.Printf("%v\n", e)
	}
}
