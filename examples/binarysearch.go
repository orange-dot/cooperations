package main

import (
	"fmt"
)

// BinarySearch implements a binary search algorithm.
// It takes a sorted slice of integers and the target value to find.
// Returns the index of the target if found, or -1 if not found.
func BinarySearch(nums []int, target int) int {
	left, right := 0, len(nums)-1

	for left <= right {
		mid := left + (right-left)/2 // Prevent overflow
		if nums[mid] == target {
			return mid
		}
		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return -1 // Target not found
}

func main() {
	nums := []int{1, 3, 5, 7, 9}
	target := 7

	result := BinarySearch(nums, target)

	fmt.Printf("Index of %d: %d\n", target, result)
	// Expected output: Index of 7: 3
}

// NEXT: reviewer