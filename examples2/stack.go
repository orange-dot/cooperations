package main

import "fmt"

// Stack represents a LIFO data structure.
type Stack[T any] struct {
	elements []T
}

// New creates and returns a new empty stack.
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

// NewWithCapacity creates a stack with pre-allocated capacity.
func NewStackWithCapacity[T any](capacity int) *Stack[T] {
	if capacity < 0 {
		capacity = 0
	}
	return &Stack[T]{elements: make([]T, 0, capacity)}
}

// Push adds an element to the top of the stack.
func (s *Stack[T]) Push(element T) {
	s.elements = append(s.elements, element)
}

// Pop removes and returns the top element.
// Returns the element and true if successful, zero value and false if empty.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	n := len(s.elements)
	if n == 0 {
		return zero, false
	}

	top := s.elements[n-1]
	// Clear reference to avoid retaining memory for large elements.
	s.elements[n-1] = zero
	s.elements = s.elements[:n-1]
	return top, true
}

// Peek returns the top element without removing it.
// Returns the element and true if successful, zero value and false if empty.
func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	n := len(s.elements)
	if n == 0 {
		return zero, false
	}
	return s.elements[n-1], true
}

// IsEmpty returns true if the stack has no elements.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}

// Size returns the number of elements in the stack.
func (s *Stack[T]) Size() int {
	return len(s.elements)
}

// Clear removes all elements from the stack.
func (s *Stack[T]) Clear() {
	var zero T
	for i := range s.elements {
		s.elements[i] = zero
	}
	s.elements = nil
}

func main() {
	s := NewStack[int]()
	fmt.Println("Push 1, 2, 3")
	s.Push(1)
	s.Push(2)
	s.Push(3)

	fmt.Printf("Size: %d\n", s.Size())

	if v, ok := s.Peek(); ok {
		fmt.Printf("Peek: %d\n", v)
	}

	for !s.IsEmpty() {
		if v, ok := s.Pop(); ok {
			fmt.Printf("Pop: %d\n", v)
		}
	}

	fmt.Printf("IsEmpty: %v\n", s.IsEmpty())
}
