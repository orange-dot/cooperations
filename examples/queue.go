package main

import (
	"fmt"
	"sync"
)

// Queue represents a FIFO (first in, first out) data structure,
// implemented using a slice for storing elements.
// Now implemented with generics for better type safety with Go 1.18+.
type Queue[T any] struct {
	items []T
	lock  sync.RWMutex
}

// NewQueue creates and returns a new Queue.
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items: make([]T, 0),
	}
}

// Enqueue adds an item to the end of the queue.
func (q *Queue[T]) Enqueue(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items = append(q.items, item)
}

// Dequeue removes and returns the item at the front of the queue.
// If the queue is empty, the second return value is false.
func (q *Queue[T]) Dequeue() (T, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	var zeroVal T // Default zero value of T
	if len(q.items) == 0 {
		return zeroVal, false
	}
	item := q.items[0]
	q.items[0] = zeroVal // Allow GC to collect the item
	q.items = q.items[1:]
	return item, true
}

// Peek returns the item at the front of the queue without removing it.
// If the queue is empty, the second return value is false.
func (q *Queue[T]) Peek() (T, bool) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	var zeroVal T // Default zero value of T
	if len(q.items) == 0 {
		return zeroVal, false
	}
	return q.items[0], true
}

// Size returns the number of items in the queue.
func (q *Queue[T]) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return len(q.items)
}

// IsEmpty checks if the queue has no elements.
func (q *Queue[T]) IsEmpty() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return len(q.items) == 0
}

// main function to demonstrate operations of Queue
func demoQueue() {
	queue := NewQueue[int]() // Example of a queue for integers

	// Enqueue items
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	// Size of queue
	fmt.Println("Size:", queue.Size()) // Output: Size: 3

	// Peek item
	peekItem, ok := queue.Peek()
	if ok {
		fmt.Println("Peek item:", peekItem) // Output: Peek item: 1
	}

	// Dequeue items
	for !queue.IsEmpty() {
		item, ok := queue.Dequeue()
		if ok {
			fmt.Println("Dequeue item:", item)
		}
	}

	// Size of queue after dequeue operations
	fmt.Println("Size after dequeue:", queue.Size()) // Output: Size after dequeue: 0
}

// NEXT: reviewer
