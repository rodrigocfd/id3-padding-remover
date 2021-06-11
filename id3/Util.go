package id3

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
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

func (_UtilT) SynchSafeDecode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7f00_0000)
	for mask != 0 {
		out >>= 1
		out |= n & mask
		mask >>= 8
	}
	return out
}

// Parses null-separated strings, ANSI or UTF-16.
func (_UtilT) ParseAnyStrings(src []byte) ([]string, error) {
	switch src[0] {
	case 0x00:
		// Encoding is ISO-8859-1.
		return _Util.ParseAnsiStrings(src[1:]), nil // skip 0x00 encoding byte
	case 0x01:
		// Encoding is Unicode UTF-16, may have 2-byte BOM.
		return _Util.ParseUtf16Strings(src[1:]), nil // skip 0x01 encoding byte
	default:
		return nil, errors.New(
			fmt.Sprintf("Text frame with unknown text encoding (%d).", src[0]),
		)
	}
}

// Parses null-separated ASCII strings.
func (_UtilT) ParseAnsiStrings(src []byte) []string {
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
		if bom == 0xfeff || bom == 0xfffe { // BOM mark found
			if bom == 0xfffe { // we have a big-endian string, change our decoder
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
		if int(ch) > 255 {
			return false
		}
	}
	return true
}

// Serializes null-separated strings, ASCII or UTF-16.
func (_UtilT) SerializeAnyStrings(strs []string) []byte {
	isAscii := false
	totalChars := 0
	for _, str := range strs {
		if _Util.IsStringAscii(str) {
			isAscii = true
		}
		totalChars += len([]rune(str))
	}

	var blob []byte
	if isAscii {
		blob = make([]byte, totalChars+len(strs)) // include encoding byte and null separators
		blob[0] = 0x00                            // ASCII encoding
		_Util.SerializeAsciiStrings(blob[1:], strs)
	} else {
		blob = make([]byte, 1+(totalChars+len(strs))*2) // include encoding byte and null separators
		blob[0] = 0x01                                  // UTF-16 encoding
		binary.LittleEndian.PutUint16(blob[1:], 0xfeff) // 2-byte little-endian BOM
		_Util.SerializeUtf16StringsLE(blob[3:], strs)
	}

	return blob
}

// Serializes null-separated ASCII strings.
func (_UtilT) SerializeAsciiStrings(dest []byte, strs []string) {
	for idx, str := range strs {
		isLast := idx == len(strs)-1

		lenStr := len([]rune(str)) // https://stackoverflow.com/a/12668840/6923555
		charsAscii := make([]byte, lenStr)
		idx := 0
		for _, ch := range str {
			chAscii := byte(ch)
			if chAscii > 0 { // in some cases, 1st char after diacritic is zero
				charsAscii[idx] = chAscii
				idx++
			}
		}
		copy(dest, charsAscii)
		dest = dest[lenStr:]

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

		chars16 := utf16.Encode([]rune(str))
		for _, ch := range chars16 {
			binary.LittleEndian.PutUint16(dest, ch)
			dest = dest[2:]
		}

		if !isLast {
			binary.LittleEndian.PutUint16(dest, 0x0000) // null separator
			dest = dest[2:]
		}
	}
}
