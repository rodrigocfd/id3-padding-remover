package id3v2

import (
	"fmt"
	"id3fit/id3v2/util"
)

type FrameComment struct {
	_FrameBase
	lang string
	text string
}

// Constructor.
func _NewFrameComment(base _FrameBase, lang, text string) *FrameComment {
	return &FrameComment{
		_FrameBase: base,
		lang:       lang,
		text:       text,
	}
}

// Constructor.
func _ParseFrameComment(base _FrameBase, src []byte) (*FrameComment, error) {
	// Retrieve text encoding.
	if src[0] != 0x00 && src[0] != 0x01 {
		return nil, fmt.Errorf("unrecognized comment text encoding: %02x", src[0])
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
		return nil, fmt.Errorf("comment frame with multiple texts: %d", len(texts))
	}

	return _NewFrameComment(base, lang, texts[0]), nil
}

func (me *FrameComment) Lang() *string { return &me.lang }
func (me *FrameComment) Text() *string { return &me.text }

func (me *FrameComment) Serialize() ([]byte, error) {
	if len(me.lang) != 3 {
		return nil, fmt.Errorf("bad lang: %s", me.lang)
	}

	encodingByte, data := util.SerializeStrings([]string{me.text})
	totalFrameSize := 10 + 1 + 3 + 1 + len(data) // header + encodingByte + lang + sep

	header, err := me._FrameBase.serializeHeader(totalFrameSize)
	if err != nil {
		return nil, fmt.Errorf("serializing FrameComment header: %w", err)
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)    // 10-byte header
	final = append(final, encodingByte) // encoding byte goes before lang
	final = append(final, []byte(me.lang)...)
	final = append(final, 0x00) // sep for content description, which is not written here
	final = append(final, data...)

	return final, nil
}
