package main

import (
	"errors"
	"fmt"
)

// MinHeap struct has a slice that holds the array
type MinHeap struct {
	array []int
}

// Insert adds an element to the heap
func (h *MinHeap) Insert(key int) {
	h.array = append(h.array, key)
	h.heapifyUp(len(h.array) - 1)
}

// ExtractMin removes and returns the smallest element in the heap
func (h *MinHeap) ExtractMin() (int, error) {
	if len(h.array) == 0 {
		return 0, errors.New("heap is empty")
	}
	min := h.array[0]
	// Move the last element to the root
	h.array[0] = h.array[len(h.array)-1]
	h.array = h.array[:len(h.array)-1]
	h.heapifyDown(0)
	return min, nil
}

// heapifyUp maintains the heap property after insertion
func (h *MinHeap) heapifyUp(index int) {
	for index > 0 && h.array[parent(index)] > h.array[index] {
		h.swap(parent(index), index)
		index = parent(index)
	}
}

// heapifyDown maintains the heap property after extracting the min
func (h *MinHeap) heapifyDown(index int) {
	if len(h.array) == 0 {
		return
	}
	lastIndex := len(h.array) - 1
	l, r := left(index), right(index)
	childToCompare := 0
	for l <= lastIndex {
		if l == lastIndex { // When left child is the only child
			childToCompare = l
		} else if h.array[l] < h.array[r] {
			childToCompare = l
		} else {
			childToCompare = r
		}
		if h.array[index] > h.array[childToCompare] {
			h.swap(index, childToCompare)
			index = childToCompare
			l, r = left(index), right(index)
		} else {
			return
		}
	}
}

// Peek returns the smallest element without removing it
func (h *MinHeap) Peek() (int, error) {
	if len(h.array) == 0 {
		return 0, errors.New("heap is empty")
	}
	return h.array[0], nil
}

// IsEmpty checks if the heap is empty
func (h *MinHeap) IsEmpty() bool {
	return len(h.array) == 0
}

// Private helper functions
func parent(i int) int {
	return (i - 1) / 2
}

func left(i int) int {
	return 2*i + 1
}

func right(i int) int {
	return 2*i + 2
}

func (h *MinHeap) swap(i1, i2 int) {
	h.array[i1], h.array[i2] = h.array[i2], h.array[i1]
}

// A utility function to print the heap
func (h *MinHeap) Print() {
	fmt.Println(h.array)
}

func demoHeap() {
	minHeap := &MinHeap{}
	fmt.Println("Creating a new MinHeap...")
	nums := []int{3, 2, 1, 7, 8, 4, 10, 16, 12}
	for _, num := range nums {
		minHeap.Insert(num)
		minHeap.Print()
	}
	min, err := minHeap.ExtractMin()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("The min value is: ", min)
	minHeap.Print()
}

// NEXT: reviewer
