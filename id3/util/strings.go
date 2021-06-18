package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/rodrigocfd/windigo/win"
)

const _BOM_LE uint16 = 0xfeff
const _BOM_BE uint16 = 0xfffe

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
		return nil, errors.New(
			fmt.Sprintf("Unrecognized text encoding: %02x.", src[0]))
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

	strBlocks := bytes.Split(src, []byte{0x00})
	texts := make([]string, 0, len(strBlocks))

	for _, block := range strBlocks {
		endianDecoder, block := _GetDecoderFromBom(block)
		runes := make([]rune, 0, len(block)/2)
		for i := 0; i < len(block); i += 2 {
			runes = append(runes, rune(endianDecoder.Uint16(block[i:]))) // raw conversion
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
	for _, oneString := range theStrings {
		runeArr := []rune(oneString)
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
		blob = win.Bytes.Append16(blob, binary.LittleEndian, _BOM_LE) // encode all strings as little-endian
	}

	for _, oneString := range theStrings {
		runeArr := []rune(oneString)

		for _, ch := range runeArr {
			if isUnicode {
				blob = win.Bytes.Append16(blob, binary.LittleEndian, uint16(ch))
			} else {
				blob = append(blob, byte(ch))
			}
		}

		if isUnicode { // all strings are null-terminated
			blob = win.Bytes.Append16(blob, binary.LittleEndian, 0x0000)
		} else {
			blob = append(blob, 0x00)
		}
	}

	return
}

func _GetDecoderFromBom(src []byte) (binary.ByteOrder, []byte) {
	// Unicode strings should always start with BOM, but just in case we have a
	// faulty one, use little-endian as default.
	var endianDecoder binary.ByteOrder = binary.LittleEndian

	bom := binary.LittleEndian.Uint16(src)
	if bom == _BOM_LE || bom == _BOM_BE { // BOM mark found
		if bom == _BOM_BE { // we have a big-endian string, change our decoder
			endianDecoder = binary.BigEndian
		}
		src = src[2:] // skip BOM
	}

	return endianDecoder, src
}
