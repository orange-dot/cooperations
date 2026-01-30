package main

import (
	"fmt"
	"sync"
)

// Queue defines the interface for a FIFO queue.
type Queue[T any] interface {
	Enqueue(item T)
	Dequeue() (T, bool)
	Peek() (T, bool)
	Len() int
	IsEmpty() bool
	Clear()
}

// SliceQueue is a non-thread-safe implementation.
type SliceQueue[T any] struct {
	items []T
	head  int
}

// SafeQueue is a thread-safe implementation.
type SafeQueue[T any] struct {
	items []T
	head  int
	mu    sync.RWMutex
}

// NewQueue creates a new non-thread-safe queue.
func NewQueue[T any]() *SliceQueue[T] {
	return &SliceQueue[T]{items: make([]T, 0)}
}

// NewSafeQueue creates a new thread-safe queue.
func NewSafeQueue[T any]() *SafeQueue[T] {
	return &SafeQueue[T]{items: make([]T, 0)}
}

// SliceQueue implementation

func (q *SliceQueue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

func (q *SliceQueue[T]) Dequeue() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	if q.head >= len(q.items) {
		q.items = nil
		q.head = 0
		return zero, false
	}

	item := q.items[q.head]
	q.items[q.head] = zero
	q.head++

	// Compact occasionally to avoid unbounded slice growth.
	if q.head > 1024 && q.head*2 >= len(q.items) {
		q.items = append([]T(nil), q.items[q.head:]...)
		q.head = 0
	} else if q.head >= len(q.items) {
		q.items = nil
		q.head = 0
	}

	return item, true
}

func (q *SliceQueue[T]) Peek() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	if q.head >= len(q.items) {
		q.items = nil
		q.head = 0
		return zero, false
	}
	return q.items[q.head], true
}

func (q *SliceQueue[T]) Len() int {
	if q.head >= len(q.items) {
		return 0
	}
	return len(q.items) - q.head
}

func (q *SliceQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

func (q *SliceQueue[T]) Clear() {
	var zero T
	for i := range q.items {
		q.items[i] = zero
	}
	q.items = nil
	q.head = 0
}

// SafeQueue implementation

func (q *SafeQueue[T]) Enqueue(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

func (q *SafeQueue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	if q.head >= len(q.items) {
		q.items = nil
		q.head = 0
		return zero, false
	}

	item := q.items[q.head]
	q.items[q.head] = zero
	q.head++

	if q.head > 1024 && q.head*2 >= len(q.items) {
		q.items = append([]T(nil), q.items[q.head:]...)
		q.head = 0
	} else if q.head >= len(q.items) {
		q.items = nil
		q.head = 0
	}

	return item, true
}

func (q *SafeQueue[T]) Peek() (T, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	if q.head >= len(q.items) {
		return zero, false
	}
	return q.items[q.head], true
}

func (q *SafeQueue[T]) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.head >= len(q.items) {
		return 0
	}
	return len(q.items) - q.head
}

func (q *SafeQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

func (q *SafeQueue[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	var zero T
	for i := range q.items {
		q.items[i] = zero
	}
	q.items = nil
	q.head = 0
}

func demoQueue() {
	// Demo: SliceQueue
	fmt.Println("=== SliceQueue ===")
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	fmt.Printf("Len: %d\n", q.Len())

	if v, ok := q.Peek(); ok {
		fmt.Printf("Peek: %d\n", v)
	}

	for !q.IsEmpty() {
		if v, ok := q.Dequeue(); ok {
			fmt.Printf("Dequeue: %d\n", v)
		}
	}

	// Demo: SafeQueue
	fmt.Println("\n=== SafeQueue (thread-safe) ===")
	sq := NewSafeQueue[string]()
	sq.Enqueue("a")
	sq.Enqueue("b")
	sq.Enqueue("c")

	fmt.Printf("Len: %d\n", sq.Len())

	for !sq.IsEmpty() {
		if v, ok := sq.Dequeue(); ok {
			fmt.Printf("Dequeue: %s\n", v)
		}
	}
}
