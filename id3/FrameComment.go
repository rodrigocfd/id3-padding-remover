package id3

import (
	"errors"
	"fmt"
	"id3fit/id3/util"
)

type FrameComment struct {
	_FrameBase
	lang string
	text string
}

// Constructor.
func _ParseFrameComment(base _FrameBase, src []byte) (*FrameComment, error) {
	// Retrieve text encoding.
	if src[0] != 0x00 && src[0] != 0x01 {
		return nil, errors.New(
			fmt.Sprintf("Unrecognized comment text encoding: %02x.", src[0]))
	}
	isUnicode := src[0] == 0x01
	src = src[1:] // skip encoding byte

	// Retrieve 3-char language string, always ASCII.
	lang := string(src[:3])
	src = src[3:]

	if src[0] == 0x00 {
		src = src[1:] // a null separator may appear, skip it
	}

	// Retrieve comment text.
	var texts []string
	if isUnicode {
		texts = util.ParseUnicodeStrings(src)
	} else {
		texts = util.ParseIso88591Strings(src)
	}

	if len(texts) > 1 {
		return nil, errors.New(
			fmt.Sprintf("Comment frame with multiple texts: %d.", len(texts)))
	}

	return &FrameComment{
		_FrameBase: base,
		lang:       lang,
		text:       texts[0],
	}, nil
}

func (me *FrameComment) Lang() *string { return &me.lang }
func (me *FrameComment) Text() *string { return &me.text }

func (me *FrameComment) Serialize() []byte {
	encodingByte, data := util.SerializeStrings([]string{me.text})
	totalFrameSize := 10 + 1 + 3 + len(data) // header + encodingByte + lang

	header := me._FrameBase.serializeHeader(totalFrameSize)

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)
	final = append(final, encodingByte)
	final = append(final, []byte(me.lang)...)
	final = append(final, 0x00)
	final = append(final, data...)

	return final
}
