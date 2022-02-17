package util

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Appends an uint16 onto a []byte, returning the newly allocated slice.
func Append16(dest []byte, encoding binary.ByteOrder, val uint16) []byte {
	buf := [2]byte{}
	encoding.PutUint16(buf[:], val)
	return append(dest, buf[:]...)
}

// Appends an uint32 onto a []byte, returning the newly allocated slice.
func Append32(dest []byte, encoding binary.ByteOrder, val uint32) []byte {
	buf := [4]byte{}
	encoding.PutUint32(buf[:], val)
	return append(dest, buf[:]...)
}

// Finds the MP3 signature in the slice, if present.
func FindMp3Signature(src []byte) (int, bool) {
	for i, b := range src {
		// https://stackoverflow.com/a/7302482/6923555
		if b == 0xff && src[i+1] == 0xfb {
			return i, true
		}
	}
	return 0, false
}

// Returns the index of the given element within the slice, or -1.
func Index16(src []uint16, elemToFind uint16) int {
	for idx, word := range src {
		if word == elemToFind {
			return idx
		}
	}
	return -1
}

// Tells whether the slice contains only zeros.
func IsSliceZeroed(blob []byte) bool {
	for _, b := range blob {
		if b != 0x00 {
			return false
		}
	}
	return true
}

// Uses unsafe.Slice() to cast a []byte into a []uint16 over the same memory
// location.
func Slice8To16(src []byte) []uint16 {
	if len(src)%2 != 0 {
		panic(fmt.Sprintf(
			"Byte slice cannot be converted into uint16: %d elements.", len(src)))
	}
	return unsafe.Slice((*uint16)(unsafe.Pointer(&src[0])), len(src)/2)
}

// Uses unsafe.Slice() to cast a []uint16 into a []byte over the same memory
// location.
func Slice16To8(src []uint16) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&src[0])), len(src)*2)
}

// Splits the given slice into subslices over the same memory location.
func Split16(src []uint16, separator uint16) [][]uint16 {
	chunks := make([][]uint16, 0, 4) // arbitrary
	for {
		sepIdx := Index16(src, separator)
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
func TrimRightZeros8(src []byte) []byte {
	for len(src) > 0 && src[len(src)-1] == 0x00 {
		src = src[:len(src)-1]
	}
	return src
}

// Returns a subslice with all the zeros at the end removed.
func TrimRightZeros16(src []uint16) []uint16 {
	for len(src) > 0 && src[len(src)-1] == 0x0000 {
		src = src[:len(src)-1]
	}
	return src
}
