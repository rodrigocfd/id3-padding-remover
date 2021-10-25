package id3v2

import (
	"fmt"
	"id3fit/id3v2/util"
)

type FrameText struct {
	_FrameHeader
	text string
}

// Constructor.
func _NewFrameText(header _FrameHeader, text string) *FrameText {
	return &FrameText{
		_FrameHeader: header,
		text:         text,
	}
}

func (me *FrameText) Text() *string { return &me.text }

func (me *FrameText) Serialize() ([]byte, error) {
	encodingByte, data := util.SerializeStrings([]string{me.text})
	totalFrameSize := 10 + 1 + len(data) // header + encodingByte

	headerBlob, err := me._FrameHeader.serialize(totalFrameSize)
	if err != nil {
		return nil, fmt.Errorf("serializing FrameText header: %w", err)
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, headerBlob...) // 10-byte header
	final = append(final, encodingByte)
	final = append(final, data...)

	return final, nil
}
