package main

import (
	"fmt"
)

// mergeSort takes an array of integers and returns a sorted array using the merge sort algorithm.
func mergeSort(items []int) []int {
	if len(items) < 2 {
		return items
	}

	mid := len(items) / 2
	left := mergeSort(items[:mid])
	right := mergeSort(items[mid:])

	return merge(left, right)
}

// merge combines two sorted slices into a single, sorted slice.
func merge(left, right []int) []int {
	var result []int
	leftIndex, rightIndex := 0, 0

	for leftIndex < len(left) && rightIndex < len(right) {
		if left[leftIndex] < right[rightIndex] {
			result = append(result, left[leftIndex])
			leftIndex++
		} else {
			result = append(result, right[rightIndex])
			rightIndex++
		}
	}

	// Append any remaining elements from either left or right slice (only one will have elements)
	result = append(result, left[leftIndex:]...)
	result = append(result, right[rightIndex:]...)

	return result
}

func demoMergeSort() {
	unsorted := []int{9, 4, 3, 5, 1, 8, 7, 2, 6}
	fmt.Println("Unsorted:", unsorted)

	sorted := mergeSort(unsorted)
	fmt.Println("Sorted:", sorted)
}

// NEXT: reviewer
