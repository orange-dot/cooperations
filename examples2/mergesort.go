package main

import (
	"fmt"
	"math/rand"
	"time"
)

func MergeSort(nums []int) []int {
	if len(nums) <= 1 {
		out := make([]int, len(nums))
		copy(out, nums)
		return out
	}

	mid := len(nums) / 2
	left := MergeSort(nums[:mid])
	right := MergeSort(nums[mid:])

	return merge(left, right)
}

func merge(a, b []int) []int {
	out := make([]int, 0, len(a)+len(b))

	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}

	out = append(out, a[i:]...)
	out = append(out, b[j:]...)
	return out
}

// MergeSortInPlace sorts nums in ascending order using mergesort.
// It allocates a single scratch buffer of the same size as nums and reuses it
// throughout recursion to avoid repeated allocations.
func MergeSortInPlace(nums []int) error {
	if nums == nil {
		return nil
	}
	if len(nums) <= 1 {
		return nil
	}

	buf := make([]int, len(nums))
	mergeSortRange(nums, buf, 0, len(nums))
	return nil
}

func mergeSortRange(nums, buf []int, lo, hi int) {
	if hi-lo <= 1 {
		return
	}

	mid := lo + (hi-lo)/2
	mergeSortRange(nums, buf, lo, mid)
	mergeSortRange(nums, buf, mid, hi)
	mergeRange(nums, buf, lo, mid, hi)
}

func mergeRange(nums, buf []int, lo, mid, hi int) {
	i, j, k := lo, mid, lo

	for i < mid && j < hi {
		if nums[i] <= nums[j] {
			buf[k] = nums[i]
			i++
		} else {
			buf[k] = nums[j]
			j++
		}
		k++
	}

	for i < mid {
		buf[k] = nums[i]
		i++
		k++
	}

	for j < hi {
		buf[k] = nums[j]
		j++
		k++
	}

	copy(nums[lo:hi], buf[lo:hi])
}

func demoMergeSort() {
	rand.Seed(time.Now().UnixNano())

	nums := make([]int, 20)
	for i := range nums {
		nums[i] = rand.Intn(100)
	}

	fmt.Println("original:", nums)

	sorted := MergeSort(nums)
	fmt.Println("MergeSort (new slice):", sorted)
	fmt.Println("after MergeSort, original unchanged:", nums)

	if err := MergeSortInPlace(nums); err != nil {
		panic(err)
	}
	fmt.Println("MergeSortInPlace (in-place):", nums)
}
