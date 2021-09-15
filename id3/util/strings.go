package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
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
		return nil, fmt.Errorf("Unrecognized text encoding: %02x.", src[0])
	}
}

func ParseIso88591Strings(src []byte) []string {
	strBlocks := bytes.Split(src, []byte{0x00})
	texts := make([]string, 0, len(strBlocks))

	for _, block := range strBlocks {
		runes := make([]rune, 0, len(block))
		for _, ch := range block {
			runes = append(runes, rune(ch)) // convert byte to rune
		}
		parsedText := string(runes) // then convert []rune to string
		if parsedText != "" {
			texts = append(texts, parsedText) // only non-empty strings
		}
	}

	return texts
}

func ParseUnicodeStrings(src []byte) []string {
	if len(src)&1 != 0 {
		// Length is not even, something is not quite right.
		// We'll simply discard the last byte and hope for the best.
		src = src[:len(src)-1]
	}

	src16 := unsafe.Slice((*uint16)(unsafe.Pointer(&src[0])), len(src)/2)
	strBlocks16 := Split16(src16, 0x0000)
	texts := make([]string, 0, len(strBlocks16))

	for _, block16 := range strBlocks16 {
		// Unicode strings should always start with BOM, but just in case we
		// have a faulty one, use little-endian as default.
		var endianDecoder binary.ByteOrder = binary.LittleEndian
		if block16[0] == _BOM_LE || block16[0] == _BOM_BE {
			if block16[0] == _BOM_BE { // we have a big-endian string, change our decoder
				endianDecoder = binary.BigEndian
			}
			block16 = block16[1:] // skip BOM
		}

		runes := make([]rune, 0, len(block16))
		block8 := unsafe.Slice((*uint8)(unsafe.Pointer(&block16[0])), len(block16)*2)

		for i := 0; i < len(block8); i += 2 {
			runes = append(runes, rune(endianDecoder.Uint16(block8[i:]))) // raw conversion
		}
		parsedText := string(runes) // then convert []rune to string
		if parsedText != "" {
			texts = append(texts, parsedText) // only non-empty strings
		}
	}

	return texts
}

func SerializeStrings(theStrings []string) (encodingByte byte, blob []byte) {
	isUnicode := false
	estimatedLenBytes := 0

out:
	for _, oneString := range theStrings { // just to check if it will be Unicode
		runeArr := []rune(oneString) // convert to rune slice
		estimatedLenBytes += len(runeArr)

		for _, ch := range runeArr {
			if ch > 255 {
				isUnicode = true
				break out
			}
		}
	}

	if isUnicode {
		encodingByte = 0x01
		estimatedLenBytes *= 2                   // will store as uint16
		estimatedLenBytes += 2 * len(theStrings) // all strings are null-terminated
		estimatedLenBytes += 2                   // BOM bytes
	} else {
		encodingByte = 0x00
		estimatedLenBytes += len(theStrings) // all strings are null-terminated
	}

	blob = make([]byte, 0, estimatedLenBytes)

	if isUnicode {
		// Append the BOM bytes.
		// All strings will be encoded as little-endian.
		blob = Append16(blob, binary.LittleEndian, _BOM_LE)
	}

	for _, oneString := range theStrings {
		runeArr := []rune(oneString) // convert to rune slice

		for _, ch := range runeArr { // append each character to final blob
			if isUnicode {
				blob = Append16(blob, binary.LittleEndian, uint16(ch))
			} else {
				blob = append(blob, byte(ch))
			}
		}

		if isUnicode { // all strings are null-terminated
			blob = Append16(blob, binary.LittleEndian, 0x0000)
		} else {
			blob = append(blob, 0x00)
		}
	}

	return
}
