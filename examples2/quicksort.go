package main

import (
	"fmt"
)

// QuickSort sorts a slice of ints in ascending order in-place.
func QuickSort(a []int) {
	if len(a) < 2 {
		return
	}
	quickSort(a, 0, len(a)-1)
}

func quickSort(a []int, lo, hi int) {
	for lo < hi {
		p := partition(a, lo, hi)

		// Recurse on smaller side first to keep stack depth O(log n) on average.
		if p-lo < hi-p {
			quickSort(a, lo, p-1)
			lo = p + 1
		} else {
			quickSort(a, p+1, hi)
			hi = p - 1
		}
	}
}

func partition(a []int, lo, hi int) int {
	// Median-of-three pivot selection helps avoid worst-case on sorted inputs.
	mid := lo + (hi-lo)/2
	pivotIdx := medianOfThreeIndex(a, lo, mid, hi)
	a[pivotIdx], a[hi] = a[hi], a[pivotIdx]
	pivot := a[hi]

	i := lo
	for j := lo; j < hi; j++ {
		if a[j] < pivot {
			a[i], a[j] = a[j], a[i]
			i++
		}
	}
	a[i], a[hi] = a[hi], a[i]
	return i
}

func medianOfThreeIndex(a []int, i, j, k int) int {
	ai, aj, ak := a[i], a[j], a[k]

	if ai < aj {
		if aj < ak {
			return j
		}
		if ai < ak {
			return k
		}
		return i
	}

	// ai >= aj
	if ai < ak {
		return i
	}
	if aj < ak {
		return k
	}
	return j
}

func main() {
	data := []int{9, 3, 7, 1, 1, 2, 8, 5, 4, 6, 0}
	fmt.Println("before:", data)
	QuickSort(data)
	fmt.Println("after: ", data)
}