package main

// Stack represents a generic LIFO (Last-In-First-Out) data structure.
type Stack[T any] struct {
	items []T
}

// New creates and returns a new empty Stack.
func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

// NewWithCapacity creates a new Stack with pre-allocated capacity.
func NewWithCapacity[T any](capacity int) *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0, capacity),
	}
}

// Push adds an element to the top of the stack.
func (s *Stack[T]) Push(item T) {
	if s == nil {
		return
	}
	s.items = append(s.items, item)
}

// Pop removes and returns the top element of the stack.
// Returns the zero value and false if the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	if s == nil || len(s.items) == 0 {
		var zero T
		return zero, false
	}

	topIdx := len(s.items) - 1
	item := s.items[topIdx]
	var zero T
	s.items[topIdx] = zero
	s.items = s.items[:topIdx]

	return item, true
}

// Peek returns the top element of the stack without removing it.
// Returns the zero value and false if the stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	if s == nil || len(s.items) == 0 {
		var zero T
		return zero, false
	}

	return s.items[len(s.items)-1], true
}

// IsEmpty returns true if the stack is empty.
func (s *Stack[T]) IsEmpty() bool {
	return s == nil || len(s.items) == 0
}

// Len returns the number of elements in the stack.
func (s *Stack[T]) Len() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// Clear resets the stack to be empty.
func (s *Stack[T]) Clear() {
	if s == nil {
		return
	}
	s.items = s.items[:0]
}
