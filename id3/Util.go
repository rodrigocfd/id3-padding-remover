package id3

import (
	"encoding/binary"
)

type _UtilT struct{}

// Tag utilities.
var _Util _UtilT

func (_UtilT) IsSliceZeroed(blob []byte) bool {
	for _, b := range blob {
		if b != 0x00 {
			return false
		}
	}
	return true
}

func (_UtilT) SynchSafeEncode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7F)
	for (mask ^ 0x7FFFFFFF) != 0 {
		out = n & ^mask
		out <<= 1
		out |= n & mask
		mask = ((mask + 1) << 8) - 1
		n = out
	}
	return out
}

func (_UtilT) SynchSafeDecode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7F000000)
	for mask != 0 {
		out >>= 1
		out |= n & mask
		mask >>= 8
	}
	return out
}

// Parses null-separated ASCII strings.
func (_UtilT) ParseAsciiStrings(src []byte) []string {
	texts := make([]string, 0) // strings to be returned

	if len(src) == 0 { // no data to be parsed
		return texts
	}

	if src[len(src)-1] == 0x00 {
		src = src[:len(src)-1] // we have a trailing zero, which is useless, discard
	}

	off := 0
	for {
		if off == len(src) || src[off] == 0x00 { // we reached the end of contents, or string
			runes := make([]rune, 0, off)
			for _, ch := range src[:off] {
				runes = append(runes, rune(ch)) // convert byte to rune
			}
			parsedText := string(runes) // then convert []rune to string
			if parsedText != "" {
				texts = append(texts, parsedText) // only non-empty strings
			}

			if off == len(src) { // no more contents, we reached end of data
				break
			}
			src = src[off+1:] // skip null separator between strings
			off = 0
		} else {
			off++
		}
	}
	return texts
}

// Parses null-separated UTF-16 strings.
func (_UtilT) ParseUtf16Strings(src []byte) []string {
	if len(src)&1 != 0 {
		// Length is not even, something is not quite right.
		// We'll simply discard the last byte and hope for the best.
		src = src[:len(src)-1]
	}

	if binary.BigEndian.Uint16(src[len(src)-2:]) == 0x0000 {
		// We have a trailing zero, which is useless, discard it.
		src = src[:len(src)-2]
	}

	texts := make([]string, 0, 1) // strings to be returned
	if len(src) == 0 {
		return texts // no data to be parsed
	}

	for {
		// Strings should always start with BOM, but just in case we have a faulty
		// one, use little-endian as default.
		var endianDecoder binary.ByteOrder = binary.LittleEndian

		bom := binary.LittleEndian.Uint16(src)
		if bom == 0xFEFF || bom == 0xFFFE { // BOM mark found
			if bom == 0xFFFE { // we have a big-endian string, change our decoder
				endianDecoder = binary.BigEndian
			}
			src = src[2:] // skip BOM
		}

		off := 0
		for {
			if off == len(src) { // passed the end of data
				break
			}
			ch := endianDecoder.Uint16(src[off:])
			if ch == 0x0000 { // we found a null separator
				break
			}
			off += 2
		}

		chunk := src[:off]
		runes := make([]rune, 0, off/2)
		for i := 0; i < len(chunk); i += 2 {
			runes = append(runes, rune(endianDecoder.Uint16(chunk[i:]))) // raw conversion
		}
		texts = append(texts, string(runes)) // append parsed string

		if off == len(src) { // no more data to parse
			break
		}
		src = src[off+2:] // skip null separator
	}

	return texts
}

// Tells if a string can be serialized as ASCII, otherwise must be UTF-16.
func (_UtilT) IsStringAscii(s string) bool {
	for _, ch := range s {
		if ch > 255 {
			return false // this character can't be serialized as ASCII
		}
	}
	return true
}

// Serializes null-separated ASCII strings.
func (_UtilT) SerializeAsciiStrings(dest []byte, strs []string) {
	for idx, str := range strs {
		isLast := idx == len(strs)-1

		copy(dest, []byte(str))
		dest = dest[len(str):]

		if !isLast {
			dest[0] = 0x00 // null separator
			dest = dest[1:]
		}
	}
}

// Serializes null-separated UTF-16 strings, little-endian.
func (_UtilT) SerializeUtf16StringsLE(dest []byte, strs []string) {
	for idx, str := range strs {
		isLast := idx == len(strs)-1

		for _, ch := range str {
			binary.LittleEndian.PutUint16(dest, uint16(ch))
			dest = dest[2:]
		}

		if !isLast {
			binary.LittleEndian.PutUint16(dest, 0x0000) // null separator
			dest = dest[2:]
		}
	}
}
