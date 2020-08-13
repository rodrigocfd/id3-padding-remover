package id3

import "encoding/binary"

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
func (_UtilT) ConvertAsciiStrings(src []byte) []string {
	texts := make([]string, 0) // strings to be returned
	if len(src) == 0 {         // no data to be parsed
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
func (_UtilT) ConvertUtf16Strings(src []byte) []string {
	var endianDecoder binary.ByteOrder = binary.LittleEndian // decode text as little-endian by default
	bomMark := binary.LittleEndian.Uint16(src)
	if bomMark == 0xFEFF || bomMark == 0xFFFE { // BOM mark found
		if bomMark == 0xFFFE { // we have a big-endian string
			endianDecoder = binary.BigEndian
		}
		src = src[2:] // skip BOM
	}

	if len(src)&1 != 0 {
		// Length is not even, something is not quite right.
		// Sometimes a lonely leading zero can show up, so we simply discard.
		src = src[:len(src)-1]
	}

	texts := make([]string, 0) // strings to be returned
	if len(src) == 0 {
		return texts // no data to be parsed
	}

	wsrc := make([]uint16, 0, len(src)/2) // convert []byte to []uint16
	for len(src) > 0 {
		wsrc = append(wsrc, endianDecoder.Uint16(src)) // observe endianness
		src = src[2:]
	}

	if wsrc[len(wsrc)-1] == 0x0000 {
		wsrc = wsrc[:len(wsrc)-1] // we have a trailing zero, which is useless, discard
	}

	off := 0
	for {
		if off == len(wsrc) || wsrc[off] == 0x0000 { // we reached the end of contents, or string
			runes := make([]rune, 0, off)
			for _, ch := range wsrc[:off] {
				runes = append(runes, rune(ch)) // convert uint16 to rune
			}
			parsedText := string(runes) // then convert []rune to string
			if parsedText != "" {
				texts = append(texts, parsedText) // only non-empty strings
			}

			if off == len(wsrc) { // no more contents, we reached end of data
				break
			}
			wsrc = wsrc[off+1:] // skip null separator between strings
			off = 0
		} else {
			off++
		}
	}
	return texts
}
