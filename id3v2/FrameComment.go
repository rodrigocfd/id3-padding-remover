package id3v2

import (
	"fmt"
	"id3fit/id3v2/util"
)

type FrameComment struct {
	_FrameHeader
	lang  string
	descr string
	text  string
}

// Constructor.
func _FrameCommentNew(header _FrameHeader, lang, descr, text string) *FrameComment {
	return &FrameComment{
		_FrameHeader: header,
		lang:         lang,
		descr:        descr,
		text:         text,
	}
}

// Constructor.
func _FrameCommentParse(header _FrameHeader, src []byte) (*FrameComment, error) {
	// Retrieve text encoding.
	if src[0] != 0x00 && src[0] != 0x01 {
		return nil, fmt.Errorf("unrecognized comment text encoding: %02x", src[0])
	}
	isUnicode := src[0] == 0x01
	src = src[1:] // skip encoding byte

	// Retrieve 3-char language string, always ASCII.
	lang := string(src[:3])
	src = src[3:]

	// Retrieve comment text.
	var texts []string
	if isUnicode {
		texts = util.ParseUnicodeStrings(src)
	} else {
		texts = util.ParseIso88591Strings(src)
	}

	if len(texts) == 2 {
		return _FrameCommentNew(header, lang, texts[0], texts[1]), nil
	} else if len(texts) == 1 {
		return _FrameCommentNew(header, lang, "", texts[0]), nil
	} else {
		return nil, fmt.Errorf("comment frame with multiple texts: %d", len(texts))
	}
}

func (me *FrameComment) Lang() *string { return &me.lang }
func (me *FrameComment) Text() *string { return &me.text }

func (me *FrameComment) Serialize() ([]byte, error) {
	if len(me.lang) != 3 {
		return nil, fmt.Errorf("bad lang: %s", me.lang)
	}

	encodingByte, textsBlob := util.SerializeStrings([]string{me.descr, me.text})
	totalFrameSize := 10 + 1 + 3 + len(textsBlob) // header + encodingByte + lang + texts

	headerBlob, err := me._FrameHeader.serialize(totalFrameSize)
	if err != nil {
		return nil, fmt.Errorf("serializing FrameComment header: %w", err)
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, headerBlob...) // 10-byte header
	final = append(final, encodingByte)  // encoding byte goes before lang
	final = append(final, []byte(me.lang)...)
	final = append(final, textsBlob...)

	return final, nil
}
