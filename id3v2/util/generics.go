package util

// Constraint for any integer type.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Returns the index of the given element within the slice, or -1.
func Index[T Integer](src []T, elemToFind T) int {
	for idx, word := range src {
		if word == elemToFind {
			return idx
		}
	}
	return -1
}

// Tells whether the slice contains only zeros.
func IsSliceZeroed[T Integer](blob []T) bool {
	for _, b := range blob {
		if b != 0 {
			return false
		}
	}
	return true
}

// Splits the given slice into subslices over the same memory location.
func Split[T Integer](src []T, separator T) [][]T {
	chunks := make([][]T, 0, 4) // arbitrary
	for {
		sepIdx := Index(src, separator)
		if sepIdx == -1 { // separator not found
			chunks = append(chunks, src) // last part with all remaining elements
			break
		}
		chunks = append(chunks, src[:sepIdx])
		src = src[sepIdx+1:]
	}
	return chunks
}

// Returns a subslice with all the zeros at the end removed.
func TrimRightZeros[T Integer](src []T) []T {
	for len(src) > 0 && src[len(src)-1] == 0 {
		src = src[:len(src)-1]
	}
	return src
}
