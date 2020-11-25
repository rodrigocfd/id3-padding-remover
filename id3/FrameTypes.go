package id3

import (
	"encoding/binary"
	"errors"
	"strings"
)

type FrameBinary struct {
	*_FrameBase
	binData []byte
}

// Constructor.
func _ParseFrameBinary(frBase *_FrameBase, src []byte) *FrameBinary {
	fr := FrameBinary{
		_FrameBase: frBase,
		binData:    make([]byte, len(src)),
	}

	copy(fr.binData, src) // simply store bytes
	return &fr
}

func (me *FrameBinary) BinData() []byte { return me.binData }

func (me *FrameBinary) Serialize() []byte {
	totalFrameSize := 10 + len(me.binData)
	blob := make([]byte, 0, totalFrameSize)
	blob = append(me._FrameBase.serializeHeader(totalFrameSize))
	blob = append(blob, me.binData...)
	return blob
}

//------------------------------------------------------------------------------

type FrameComment struct {
	*_FrameBase
	lang string
	text string
}

// Constructor.
func _ParseFrameComment(frBase *_FrameBase, src []byte) (*FrameComment, error) {
	// Retrieve text encoding.
	if src[0] != 0x00 && src[0] != 0x01 {
		return nil, errors.New("Unknown comment encoding.")
	}
	isUtf16 := src[0] == 0x01
	src = src[1:] // skip encoding byte

	// Retrieve 3-char language string, always ASCII.
	lang := string(src[:3])
	src = src[3:]

	if src[0] == 0x00 {
		src = src[1:] // a null separator may appear, skip it
	}

	// Retrieve comment text.
	var texts []string
	if isUtf16 {
		texts = _Util.ParseUtf16Strings(src)
	} else {
		texts = _Util.ParseAnsiStrings(src)
	}

	return &FrameComment{
		_FrameBase: frBase,
		lang:       lang,
		text:       texts[0],
	}, nil
}

func (me *FrameComment) Lang() string { return me.lang }
func (me *FrameComment) Text() string { return me.text }

func (me *FrameComment) Serialize() []byte {
	isAscii := _Util.IsStringAscii(me.text)
	var blob []byte

	if isAscii {
		blob = make([]byte, 1+3+1+len(me.text))
		blob[0] = 0x00 // ASCII encoding
	} else {
		blob = make([]byte, 1+3+(2+len(me.text))*2)
		blob[0] = 0x01 // UTF-16 encoding
	}

	copy(blob[1:4], []byte(me.lang)) // 3-char lang string, always ASCII

	if isAscii {
		blob[4] = 0x00 // zero char before text
		_Util.SerializeAsciiStrings(blob[5:], []string{me.text})
	} else {
		binary.LittleEndian.PutUint16(blob[4:], 0xfeff) // 2-byte little-endian BOM
		binary.LittleEndian.PutUint16(blob[6:], 0x0000) // zero char before text
		_Util.SerializeUtf16StringsLE(blob[8:], []string{me.text})
	}

	return blob
}

//------------------------------------------------------------------------------

type FrameMultiText struct {
	*_FrameBase
	texts []string
}

// Constructor.
func _ParseFrameMultiText(frBase *_FrameBase, texts []string) *FrameMultiText {
	return &FrameMultiText{
		_FrameBase: frBase,
		texts:      texts,
	}
}

func (me *FrameMultiText) Texts() []string { return me.texts }

func (me *FrameMultiText) Serialize() []byte {
	blobStr := _Util.SerializeAnyStrings(me.texts)
	totalFrameSize := 10 + len(blobStr)
	blob := make([]byte, 0, totalFrameSize)
	blob = append(me._FrameBase.serializeHeader(totalFrameSize))
	blob = append(blob, blobStr...)
	return blob
}

func (me *FrameMultiText) IsReplayGain() bool {
	return me._FrameBase.Name4() == "TXXX" &&
		len(me.texts) == 2 &&
		(strings.HasPrefix(me.texts[0], "replaygain_") ||
			strings.HasPrefix(me.texts[1], "replaygain_"))
}

//------------------------------------------------------------------------------

type FrameText struct {
	*_FrameBase
	text string
}

// Constructor.
func _ParseFrameText(frBase *_FrameBase, texts []string) *FrameText {
	return &FrameText{
		_FrameBase: frBase,
		text:       texts[0],
	}
}

func (me *FrameText) Text() string        { return me.text }
func (me *FrameText) SetText(text string) { me.text = text }

func (me *FrameText) Serialize() []byte {
	blobStr := _Util.SerializeAnyStrings([]string{me.text})
	totalFrameSize := 10 + len(blobStr)
	blob := make([]byte, 0, totalFrameSize)
	blob = append(me._FrameBase.serializeHeader(totalFrameSize))
	blob = append(blob, blobStr...)
	return blob
}
