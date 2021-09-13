package id3

import (
	"fmt"
	"id3fit/id3/util"
)

type FrameComment struct {
	_FrameBase
	lang string
	text string
}

func (me *FrameComment) Lang() *string { return &me.lang }
func (me *FrameComment) Text() *string { return &me.text }

func (me *FrameComment) parse(base _FrameBase, src []byte) error {
	// Retrieve text encoding.
	if src[0] != 0x00 && src[0] != 0x01 {
		return fmt.Errorf("Unrecognized comment text encoding: %02x.", src[0])
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
		return fmt.Errorf("Comment frame with multiple texts: %d.", len(texts))
	}

	me._FrameBase = base
	me.lang = lang
	me.text = texts[0]
	return nil
}

func (me *FrameComment) Serialize() ([]byte, error) {
	encodingByte, data := util.SerializeStrings([]string{me.text})
	totalFrameSize := 10 + 1 + 3 + len(data) // header + encodingByte + lang

	header, err := me._FrameBase.serializeHeader(totalFrameSize)
	if err != nil {
		return nil, err
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)
	final = append(final, encodingByte)
	final = append(final, []byte(me.lang)...)
	final = append(final, 0x00)
	final = append(final, data...)

	return final, nil
}
