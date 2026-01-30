package main

import (
	"errors"
	"fmt"
)

// BinarySearch searches for target in a sorted slice of ints.
// It returns the index of target, or an error if not found.
// The input must be sorted in ascending order.
func BinarySearch(a []int, target int) (int, error) {
	lo, hi := 0, len(a)-1

	for lo <= hi {
		mid := lo + (hi-lo)/2
		v := a[mid]

		switch {
		case v == target:
			return mid, nil
		case v < target:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}

	return -1, errors.New("target not found")
}

func main() {
	data := []int{1, 3, 4, 7, 9, 12, 18}

	idx, err := BinarySearch(data, 9)
	if err != nil {
		fmt.Println("search error:", err)
		return
	}
	fmt.Println("found at index:", idx)

	_, err = BinarySearch(data, 2)
	if err != nil {
		fmt.Println("search error:", err)
	}
}
