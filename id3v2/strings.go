//go:build windows

package id3v2

import (
	"fmt"
	"id3fit/slices2"
	"math/bits"
	"slices"
	"strings"
	"unsafe"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/heap"
)

// String encoding.
type ENC byte

const (
	ENC_ISO88591 ENC = 0x00
	ENC_UNICODE  ENC = 0x01

	_BOM_BE uint16 = 0xfeff
	_BOM_LE uint16 = 0xfffe
)

// Parses one or more null-separated strings, ISO-8859-1 or Unicode.
func parseStrings(src []byte) ([]string, error) {
	switch ENC(src[0]) {
	case ENC_ISO88591:
		return parseIso88591Strings(src[1:]), nil
	case ENC_UNICODE:
		return parseUnicodeStrings(src[1:]), nil
	default:
		return nil, fmt.Errorf("unrecognized text encoding: %02x", src[0])
	}
}

// Parses one or more null-separated ISO-8859-1 strings.
func parseIso88591Strings(src []byte) []string {
	src = slices2.TrimRight(src, 0x00) // right-trim zeros to avoid an extra empty string
	if len(src) == 0 {
		return []string{} // no strings
	}

	blocks := slices.Collect(slices2.Split(src, 0x00))
	texts := make([]string, 0, len(blocks))
	wideStrBuf := heap.NewWideStr[heap.Stack20]() // buffer to convert bytes to Go strings

	for _, block := range blocks {
		if len(block) == 0 {
			texts = append(texts, "") // empty strings are also added
		} else {
			wideStrBuf.Resize(uint(len(block)))
			wideStrBuf.ZeroBuffer()

			for i, ch := range block {
				*wideStrBuf.At(uint(i)) = uint16(ch)
			}
			texts = append(texts, win.Str.FromUtf16Slice(wideStrBuf.HotSlice()))
		}
	}
	return texts
}

// Parses one or more null-separated Unicode strings.
func parseUnicodeStrings(src []byte) []string {
	if len(src)%2 != 0 {
		// Length is not even, something is not quite right.
		// Discard last byte and hope for the best.
		src = src[0 : len(src)-1]
	}

	wsrc := unsafe.Slice((*uint16)(unsafe.Pointer(&src[0])), len(src)/2)
	wsrc = slices2.TrimRight(wsrc, 0x0000) // right-trim zeros to avoid an extra empty string
	if len(wsrc) == 0 {
		return []string{} // no strings
	}

	blocks := slices.Collect(slices2.Split(wsrc, 0x0000))
	texts := make([]string, 0, len(blocks))
	wideStrBuf := heap.NewWideStr[heap.Stack20]() // buffer to convert bytes to Go strings

	for _, block := range blocks {
		isLE := true
		if block[0] == _BOM_LE || block[0] == _BOM_BE { // we have a BOM
			if block[0] == _BOM_BE {
				isLE = false
			}
			block = block[1:] // skip BOM
		}

		if len(block) == 0 {
			texts = append(texts, "") // empty strings are also added
		} else {
			wideStrBuf.Resize(uint(len(block)))
			wideStrBuf.ZeroBuffer()

			for i, ch := range block {
				if isLE {
					ch = bits.ReverseBytes16(ch)
				}
				*wideStrBuf.At(uint(i)) = ch
			}
			texts = append(texts, win.Str.FromUtf16Slice(wideStrBuf.HotSlice()))
		}
	}
	return texts
}

// Serializes the given strings as null-terminated, with the proper encoding.
func serializeStrings(strs ...string) (ENC, []byte) {
	encoding := ENC_ISO88591
	estimatedLenBytes := 0

	for _, str := range strs {
		estimatedLenBytes += len(str) + 1
		hasUnicodeCh := strings.ContainsFunc(str, func(ch rune) bool { return ch > 0xff })
		if hasUnicodeCh {
			encoding = ENC_UNICODE // at least 1 string is Unicode
		}
	}

	if encoding == ENC_UNICODE { // chars will be serialized as WORD
		estimatedLenBytes *= 2
		estimatedLenBytes += 2 * len(strs) // one BOM to each string
	}

	buf := make([]byte, 0, estimatedLenBytes) // to be returned
	str16 := heap.NewWideStr[heap.Stack20]()  // to serialize each Go string

	for _, str := range strs {
		if encoding == ENC_UNICODE {
			// Insert BOM bytes for each string.
			// Strings will be encoded as little-endian.
			buf = append(buf, win.LOBYTE(_BOM_LE), win.HIBYTE(_BOM_LE))
		}

		str16.Set(str, heap.ALLOW_EMPTY) // contains terminating null

		for _, ch := range str16.HotSlice() { // write each char of the string
			if encoding == ENC_UNICODE {
				buf = append(buf, win.LOBYTE(ch), win.HIBYTE(ch))
			} else {
				buf = append(buf, win.LOBYTE(ch))
			}
		}
	}

	return encoding, buf
}
