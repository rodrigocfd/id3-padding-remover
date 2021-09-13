package util

import (
	"encoding/binary"
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

func IsSliceZeroed(blob []byte) bool {
	for _, b := range blob {
		if b != 0x00 {
			return false
		}
	}
	return true
}

// Splits the slice into chunks over the same underlying memory block.
func Split16(src []uint16, sep uint16) [][]uint16 {
	chunks := make([][]uint16, 0, 4) // arbitrary
	beginIdx, curIdx := 0, 0

	for {
		if curIdx == len(src) || src[curIdx] == sep {
			size := curIdx - beginIdx
			if size > 0 {
				chunks = append(chunks, unsafe.Slice(&src[beginIdx], size))
			}

			for curIdx != len(src) && src[curIdx] == sep {
				curIdx++ // find the next non-separator
			}
			if curIdx == len(src) {
				break // we reached end of slice
			}

			beginIdx = curIdx

		} else {
			curIdx++
		}
	}

	return chunks
}

func SynchSafeDecode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7f00_0000)
	for mask != 0 {
		out >>= 1
		out |= n & mask
		mask >>= 8
	}
	return out
}

func SynchSafeEncode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7f)
	for (mask ^ 0x7fff_ffff) != 0 {
		out = n & ^mask
		out <<= 1
		out |= n & mask
		mask = ((mask + 1) << 8) - 1
		n = out
	}
	return out
}
