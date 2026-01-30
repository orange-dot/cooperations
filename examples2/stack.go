// go.mod
module example.com/stack

go 1.22
```

```go
// stack/stack.go
package stack

// Stack represents a LIFO data structure.
type Stack[T any] struct {
	elements []T
}

// Stacker defines the interface for stack operations.
type Stacker[T any] interface {
	Push(element T)
	Pop() (T, bool)
	Peek() (T, bool)
	IsEmpty() bool
	Size() int
	Clear()
}

// New creates and returns a new empty stack.
func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

// NewWithCapacity creates a stack with pre-allocated capacity.
func NewWithCapacity[T any](capacity int) *Stack[T] {
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
```

```go
// stack/stack_test.go
package stack

import "testing"

func TestNewStackIsEmpty(t *testing.T) {
	s := New[int]()
	if !s.IsEmpty() {
		t.Fatalf("expected empty stack")
	}
	if got := s.Size(); got != 0 {
		t.Fatalf("expected size 0, got %d", got)
	}
}

func TestPushPeekPop(t *testing.T) {
	s := New[string]()

	s.Push("a")
	s.Push("b")

	if got, ok := s.Peek(); !ok || got != "b" {
		t.Fatalf("expected peek 'b' ok=true, got %q ok=%v", got, ok)
	}

	if got, ok := s.Pop(); !ok || got != "b" {
		t.Fatalf("expected pop 'b' ok=true, got %q ok=%v", got, ok)
	}
	if got, ok := s.Pop(); !ok || got != "a" {
		t.Fatalf("expected pop 'a' ok=true, got %q ok=%v", got, ok)
	}

	if _, ok := s.Pop(); ok {
		t.Fatalf("expected pop from empty stack ok=false")
	}
	if _, ok := s.Peek(); ok {
		t.Fatalf("expected peek from empty stack ok=false")
	}
}

func TestNewWithCapacity(t *testing.T) {
	s := NewWithCapacity[int](10)
	if s == nil {
		t.Fatalf("expected non-nil stack")
	}
	if !s.IsEmpty() {
		t.Fatalf("expected empty stack")
	}
	s.Push(1)
	if got := s.Size(); got != 1 {
		t.Fatalf("expected size 1, got %d", got)
	}
}

func TestClear(t *testing.T) {
	s := New[int]()
	for i := 0; i < 5; i++ {
		s.Push(i)
	}
	if s.IsEmpty() {
		t.Fatalf("expected non-empty stack before Clear")
	}

	s.Clear()
	if !s.IsEmpty() {
		t.Fatalf("expected empty stack after Clear")
	}
	if got := s.Size(); got != 0 {
		t.Fatalf("expected size 0 after Clear, got %d", got)
	}

	// Should still be usable after Clear.
	s.Push(42)
	if got, ok := s.Peek(); !ok || got != 42 {
		t.Fatalf("expected peek 42 ok=true after Clear, got %d ok=%v", got, ok)
	}
}

func TestNewWithCapacityNegative(t *testing.T) {
	s := NewWithCapacity[int](-1)
	if s == nil {
		t.Fatalf("expected non-nil stack")
	}
	if !s.IsEmpty() || s.Size() != 0 {
		t.Fatalf("expected empty stack with size 0")
	}
}