package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/rodrigocfd/windigo/win"
)

const (
	_BOM_LE uint16 = 0xfeff
	_BOM_BE uint16 = 0xfffe
)

// Parses one or more null-separated strings, ISO-8859-1 or Unicode.
func ParseAnyStrings(src []byte) ([]string, error) {
	switch src[0] {
	case 0x00:
		// Encoding is ISO-8859-1.
		return ParseIso88591Strings(src[1:]), nil
	case 0x01:
		// Encoding is Unicode, may have 2-byte BOM.
		return ParseUnicodeStrings(src[1:]), nil
	default:
		return nil, fmt.Errorf("unrecognized text encoding: %02x", src[0])
	}
}

// Parses one or more null-separated ISO-8859-1 strings.
func ParseIso88591Strings(src []byte) []string {
	src = TrimRightZeros8(src) // avoid an extra empty string
	if len(src) == 0 {
		return []string{}
	}

	texts := make([]string, 0, 2) // arbitrary
	parts := bytes.Split(src, []byte{0x00})
	for _, part := range parts {
		if len(part) == 0 {
			texts = append(texts, "") // empty strings are also added
		} else {
			buf16 := make([]uint16, 0, len(part)+1) // room for terminating null
			for _, ch := range part {
				buf16 = append(buf16, uint16(ch))
			}
			buf16 = append(buf16, 0x0000) // terminating null
			texts = append(texts, win.Str.FromNativeSlice(buf16))
		}
	}
	return texts
}

// Parses one or more null-separated Unicode strings.
func ParseUnicodeStrings(src []byte) []string {
	if len(src)%1 != 0 {
		// Length is not even, something is not quite right.
		// Discard last byte and hope for the best.
		src = src[:len(src)-1]
	}

	src16 := Slice8To16(src)
	src16 = TrimRightZeros16(src16) // avoid an extra empty string
	if len(src16) == 0 {
		return []string{}
	}

	texts := make([]string, 0, 2) // arbitrary
	parts := Split16(src16, 0x0000)
	for _, part := range parts {
		var endianDecoder binary.ByteOrder = binary.LittleEndian // little-endian by default
		if part[0] == _BOM_LE || part[0] == _BOM_BE {
			if part[0] == _BOM_BE {
				endianDecoder = binary.BigEndian
			}
			part = part[1:] // skip BOM
		}

		if len(part) == 0 {
			texts = append(texts, "") // empty strings are also added
		} else {
			part8 := Slice16To8(part)
			buf16 := make([]uint16, 0, len(part)+1) // room for terminating null
			for i := 0; i < len(part8); i += 2 {
				buf16 = append(buf16, endianDecoder.Uint16(part8[i:]))
			}
			buf16 = append(buf16, 0x0000) // terminating null
			texts = append(texts, win.Str.FromNativeSlice(buf16))
		}
	}
	return texts
}

// Serializes the given strings as null-terminated, with the proper encoding.
func SerializeStrings(theStrings []string) (encodingByte byte, serialized []byte) {
	isUnicode := false
	estimatedLenBytes := 0

	for _, oneString := range theStrings {
		estimatedLenBytes += len(oneString) + 1 // strings will be null-terminated

		if !isUnicode {
			idxUnicodeChar := strings.IndexFunc(oneString, func(ch rune) bool { return ch > 0xff })
			if idxUnicodeChar != -1 {
				isUnicode = true
			}
		}
	}

	if isUnicode {
		estimatedLenBytes *= 2
		estimatedLenBytes += 2 * len(theStrings) // BOM bytes
	}

	buf := make([]byte, 0, estimatedLenBytes)
	for _, oneString := range theStrings {
		if isUnicode {
			// Insert BOM bytes.
			// Strings will be encoded as little-endian.
			buf = Append16(buf, binary.LittleEndian, _BOM_LE)
		}

		slice16 := win.Str.ToNativeSlice(oneString) // this slice is null-terminated
		for _, ch := range slice16 {
			if isUnicode {
				buf = Append16(buf, binary.LittleEndian, ch)
			} else {
				buf = append(buf, byte(ch))
			}
		}
	}

	if isUnicode {
		return 0x01, buf
	} else {
		return 0x00, buf
	}
}
