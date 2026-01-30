package main

import (
	"errors"
	"fmt"
)

// MinHeap is a binary min-heap for ints.
//
// This type is not thread-safe. If accessed from multiple goroutines,
// callers must provide external synchronization.
type MinHeap struct {
	data []int
}

// NewMinHeap creates an empty heap with optional initial capacity.
func NewMinHeap(capacity int) *MinHeap {
	if capacity < 0 {
		capacity = 0
	}
	return &MinHeap{data: make([]int, 0, capacity)}
}

// Len returns the number of elements in the heap.
func (h *MinHeap) Len() int { return len(h.data) }

// IsEmpty reports whether the heap is empty.
func (h *MinHeap) IsEmpty() bool { return len(h.data) == 0 }

// Clear removes all elements from the heap while retaining underlying capacity.
func (h *MinHeap) Clear() { h.data = h.data[:0] }

// Peek returns the minimum element without removing it.
func (h *MinHeap) Peek() (int, error) {
	if len(h.data) == 0 {
		return 0, errors.New("minheap: peek from empty heap")
	}
	return h.data[0], nil
}

// Push inserts a value into the heap.
func (h *MinHeap) Push(x int) {
	h.data = append(h.data, x)
	h.siftUp(len(h.data) - 1)
}

// Pop removes and returns the minimum element.
func (h *MinHeap) Pop() (int, error) {
	if len(h.data) == 0 {
		return 0, errors.New("minheap: pop from empty heap")
	}

	min := h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last]
	h.data = h.data[:last]
	if len(h.data) > 0 {
		h.siftDown(0)
	}
	return min, nil
}

// ReplaceMin replaces the minimum element with x and returns the old minimum.
// This is more efficient than Pop followed by Push.
func (h *MinHeap) ReplaceMin(x int) (int, error) {
	if len(h.data) == 0 {
		return 0, errors.New("minheap: replace on empty heap")
	}
	min := h.data[0]
	h.data[0] = x
	h.siftDown(0)
	return min, nil
}

// Heapify builds a heap from the provided slice in O(n) time.
//
// Heapify copies the input values into the heap; it does not retain a reference
// to the provided slice, so modifying the original slice after calling Heapify
// will not affect the heap.
func (h *MinHeap) Heapify(values []int) {
	h.data = append(h.data[:0], values...)
	n := len(h.data)
	if n == 0 {
		return
	}
	for i := parent(n - 1); i >= 0; i-- {
		h.siftDown(i)
	}
}

func (h *MinHeap) siftUp(i int) {
	for i > 0 {
		p := parent(i)
		if h.data[p] <= h.data[i] {
			return
		}
		h.data[p], h.data[i] = h.data[i], h.data[p]
		i = p
	}
}

func (h *MinHeap) siftDown(i int) {
	n := len(h.data)
	for {
		l := left(i)
		if l >= n {
			return
		}
		r := right(i)

		smallest := l
		if r < n && h.data[r] < h.data[l] {
			smallest = r
		}
		if h.data[i] <= h.data[smallest] {
			return
		}
		h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
		i = smallest
	}
}

func parent(i int) int { return (i - 1) / 2 }
func left(i int) int   { return 2*i + 1 }
func right(i int) int  { return 2*i + 2 }

// ---- demo ----

func main() {
	h := NewMinHeap(0)

	for _, v := range []int{5, 3, 8, 1, 2, 9, 7} {
		h.Push(v)
	}

	for !h.IsEmpty() {
		x, err := h.Pop()
		if err != nil {
			panic(err)
		}
		fmt.Print(x, " ")
	}
	fmt.Println()

	h.Heapify([]int{10, 4, 6, 3, 8})
	fmt.Println("len:", h.Len())
	top, _ := h.Peek()
	fmt.Println("peek:", top)
	replaced, _ := h.ReplaceMin(5)
	fmt.Println("replaced:", replaced)
	top, _ = h.Peek()
	fmt.Println("peek:", top)

	h.Clear()
	fmt.Println("cleared, empty:", h.IsEmpty())
}