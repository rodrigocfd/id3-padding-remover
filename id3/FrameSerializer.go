package id3

import (
	"encoding/binary"
)

type _FrameSerializerT struct{}

// Frame serializer.
var _FrameSerializer _FrameSerializerT

func (_FrameSerializerT) SerializeFrame(frame Frame) []byte {

	println("next", frame.Name4())

	blob := make([]byte, 0, 10) // header is 10 bytes
	blob = append(blob, []byte(frame.Name4())...)
	blob = append(blob, []byte{0, 0, 0, 0}...) // size not yet known
	blob = append(blob, []byte{0, 0}...)       // flags

	var data []byte

	switch myFrame := frame.(type) {
	case *FrameComment:
		data = _FrameSerializer.serializeCommentFrame(myFrame)
	case *FrameText:
		data = _FrameSerializer.serializeTextsOfFrame([]string{myFrame.Text()})
	case *FrameMultiText:
		data = _FrameSerializer.serializeTextsOfFrame(myFrame.Texts())
	case *FrameBinary:
		data = myFrame.BinData()
	}

	binary.BigEndian.PutUint32(blob[4:], uint32(len(data))) // write frame size without 10-byte header
	blob = append(blob, data...)
	return blob
}

func (_FrameSerializerT) serializeCommentFrame(frame *FrameComment) []byte {
	isAscii := _Util.IsStringAscii(frame.Text())
	var blob []byte

	if isAscii {
		blob = make([]byte, 1+3+1+len(frame.Text()))
		blob[0] = 0x00 // ASCII encoding
	} else {
		blob = make([]byte, 1+3+(2+len(frame.Text()))*2)
		blob[0] = 0x01 // UTF-16 encoding
	}

	copy(blob[1:4], []byte(frame.Lang())) // 3-char lang string, always ASCII

	if isAscii {
		blob[4] = 0x00 // zero char before text
		_Util.SerializeAsciiStrings(blob[5:], []string{frame.Text()})
	} else {
		binary.LittleEndian.PutUint16(blob[4:], 0xFEFF) // 2-byte little-endian BOM
		binary.LittleEndian.PutUint16(blob[6:], 0x0000) // zero char before text
		_Util.SerializeUtf16StringsLE(blob[8:], []string{frame.Text()})
	}

	return blob
}

func (_FrameSerializerT) serializeTextsOfFrame(strs []string) []byte {
	isAscii := false
	totalChars := 0
	for _, str := range strs {
		if _Util.IsStringAscii(str) {
			isAscii = true
			totalChars += len(str)
		}
	}

	var blob []byte
	if isAscii {
		blob = make([]byte, totalChars+len(strs)) // include encoding byte and null separators
		blob[0] = 0x00                            // ASCII encoding
		_Util.SerializeAsciiStrings(blob[1:], strs)
	} else {
		blob = make([]byte, 1+(totalChars+len(strs))*2) // include encoding byte and null separators
		blob[0] = 0x01                                  // UTF-16 encoding
		binary.LittleEndian.PutUint16(blob[1:], 0xFEFF) // 2-byte little-endian BOM
		_Util.SerializeUtf16StringsLE(blob[3:], strs)
	}

	return blob
}
