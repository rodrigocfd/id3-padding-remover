package id3

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type _FrameParserT struct{}

// Frame parser.
var _FrameParser _FrameParserT

func (_FrameParserT) ParseFrame(src []byte) (Frame, error) {
	name4 := string(src[0:4])
	totalFrameSize := binary.BigEndian.Uint32(src[4:8]) + 10 // also count 10-byte tag header

	src = src[10:totalFrameSize] // skip frame header, limit to frame size

	if name4 == "COMM" { // comment frame
		frameComm, err := _FrameParser.parseCommentFrame(src)
		if err != nil {
			return nil, err
		}
		frameComm.name4 = name4
		frameComm.totalFrameSize = uint(totalFrameSize)
		return frameComm, nil

	} else if name4[0] == 'T' {
		parsedTexts, err := _FrameParser.parseTextsOfFrame(src)
		if err != nil {
			return nil, err
		}

		if len(parsedTexts) == 0 {
			return nil, errors.New("Frame text contains no texts.")

		} else if len(parsedTexts) == 1 { // simple text frame
			frameText := &FrameText{}
			frameText.name4 = name4
			frameText.totalFrameSize = uint(totalFrameSize)
			frameText.text = parsedTexts[0]
			return frameText, nil

		} else if len(parsedTexts) > 1 { // multi text frame
			frameTexts := &FrameMultiText{}
			frameTexts.name4 = name4
			frameTexts.totalFrameSize = uint(totalFrameSize)
			frameTexts.texts = parsedTexts
			return frameTexts, nil
		}
	}

	// Anything else is treated as raw binary.
	frameBin := &FrameBinary{}
	frameBin.name4 = name4
	frameBin.totalFrameSize = uint(totalFrameSize)
	frameBin.binData = make([]byte, len(src))
	copy(frameBin.binData, src) // simply store bytes
	return frameBin, nil
}

func (_FrameParserT) parseCommentFrame(src []byte) (*FrameComment, error) {
	frameComm := &FrameComment{}

	// Retrieve text encoding.
	if src[0] != 0x00 && src[0] != 0x01 {
		return nil, errors.New("Unknown comment encoding.")
	}
	isUtf16 := src[0] == 0x01
	src = src[1:] // skip encoding byte

	// Retrieve 3-char language string, always ASCII.
	frameComm.lang = string(src[:3])
	src = src[3:]

	if src[0] == 0x00 {
		src = src[1:] // a null separator may appear, skip it
	}

	// Retrieve comment text.
	var texts []string
	if isUtf16 {
		texts = _Util.ParseUtf16Strings(src)
	} else {
		texts = _Util.ParseAsciiStrings(src)
	}

	frameComm.text = texts[len(texts)-1] // if more than one, get the last one
	return frameComm, nil
}

func (_FrameParserT) parseTextsOfFrame(src []byte) ([]string, error) {
	switch src[0] {
	case 0x00:
		// Encoding is ISO-8859-1.
		return _Util.ParseAsciiStrings(src[1:]), nil // skip 0x00 encoding byte
	case 0x01:
		// Encoding is Unicode UTF-16, may have 2-byte BOM.
		return _Util.ParseUtf16Strings(src[1:]), nil // skip 0x01 encoding byte
	default:
		return nil, errors.New(
			fmt.Sprintf("Text frame with unknown text encoding (%d).", src[0]),
		)
	}
}
