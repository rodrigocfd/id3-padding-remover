package slices2

import (
	"iter"
	"slices"
)

// Returns true if all elements are equal to the given one.
func AllEqual[T comparable](src []T, elemToCompare T) bool {
	for _, elem := range src {
		if elem != elemToCompare {
			return false
		}
	}
	return true
}

// Returns true if the predicate returns true for all elements.
func AllTrueFunc[S ~[]T, T comparable](src S, pred func(elem T) bool) bool {
	for _, elem := range src {
		if !pred(elem) {
			return false
		}
	}
	return true
}

// Returns a newly allocated slice with the contents copied.
func Clone[T any](src []T) []T {
	cloned := make([]T, len(src))
	copy(cloned, src)
	return cloned
}

// Returns the index of the last element to which the predicate returns true.
//
// # Example
//
//	nums := []uint{400, 500, 600, 700}
//
//	idx := slices2.LastIndexFunc(nums, func(n uint) bool {
//		return n == 600
//	})
func LastIndexFunc[S ~[]T, T comparable](src S, pred func(elem T) bool) int {
	for i := len(src) - 1; i >= 0; i-- {
		if pred(src[i]) {
			return i
		}
	}
	return -1
}

// Returns a new slice by mapping each element according to the callback.
//
// # Example
//
//	nums := []uint{400, 500, 600, 700}
//
//	strs := slices2.Map(nums, func(index int, num uint) string {
//		return fmt.Sprintf("Num %d", num)
//	})
func Map[T, U any](src []T, fun func(index int, elem T) U) []U {
	mapped := make([]U, 0, len(src))
	for i := range len(src) {
		mapped = append(mapped, fun(i, src[i]))
	}
	return mapped
}

// Iterates over slices over the original slice, separated by the given
// separator.
func Split[S ~[]T, T comparable](src S, separator T) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		for {
			sepIdx := slices.Index(src, separator)
			if sepIdx == -1 { // separator not found
				yield(src) // last part with all remaining elements
				break
			}
			if !yield(src[:sepIdx]) {
				break
			}
			src = src[sepIdx+1:]
		}
	}
}

// Returns a subslice without the first and last elements which contiguously
// match the given element.
func Trim[T comparable](src []T, elem T) []T {
	return TrimRight(TrimLeft(src, elem), elem)
}

// Returns a subslice without the first elements which contiguously match the
// given element.
func TrimLeft[T comparable](src []T, elem T) []T {
	for i := range len(src) {
		if src[i] != elem {
			return src[i:]
		}
	}
	return []T{}
}

// Returns a subslice without the last elements which contiguously match the
// given element.
func TrimRight[T comparable](src []T, elem T) []T {
	for i := len(src) - 1; i >= 0; i-- {
		if src[i] != elem {
			return src[:i+1]
		}
	}
	return []T{}
}
