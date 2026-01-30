package main

import (
	"fmt"
)

// quickSort sorts the given slice in-place using the QuickSort algorithm.
// It modifies the slice and returns the same slice for convenience in chaining.
func quickSort(arr []int) []int {
	if len(arr) < 2 {
		return arr
	}
	left, right := 0, len(arr)-1

	// Pick a pivot element. We're choosing the middle element.
	pivot := arr[len(arr)/2]

	for left <= right {
		for arr[left] < pivot {
			left++
		}
		// Changed while to for to fix the syntax error
		for arr[right] > pivot {
			right--
		}

		if left <= right {
			arr[left], arr[right] = arr[right], arr[left]
			left++
			right--
		}
	}

	// Recursively sort the two halves
	if right > 0 {
		quickSort(arr[:right+1])
	}
	if left < len(arr) {
		quickSort(arr[left:])
	}

	return arr
}

func demoQuickSort() {
	testCases := [][]int{
		{9, 3, 4, 2, 1, 8, 5, 7, 6}, // Typical unsorted array
		{},                          // Empty array
		{1},                         // Single element
		{2, 1},                      // Two elements
		{1, 1, 1, 1},                // All elements the same
		{1, 2, 3, 4, 5},             // Already sorted
		{5, 4, 3, 2, 1},             // Reverse sorted
	}

	for _, arr := range testCases {
		fmt.Printf("Input: %v -> ", arr)
		quickSort(arr)
		fmt.Printf("Sorted: %v\n", arr)
	}
}
