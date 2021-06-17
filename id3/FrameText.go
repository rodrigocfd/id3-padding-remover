package id3

import (
	"id3fit/id3/util"
)

type FrameText struct {
	_FrameBase
	text string
}

// Constructor.
func _ParseFrameText(base _FrameBase, texts []string) (*FrameText, error) {
	return &FrameText{
		_FrameBase: base,
		text:       texts[0],
	}, nil
}

func (me *FrameText) Text() *string { return &me.text }

func (me *FrameText) Serialize() []byte {
	encodingByte, data := util.SerializeStrings([]string{me.text})
	totalFrameSize := 10 + 1 + len(data) // header + encodingByte

	header := me._FrameBase.serializeHeader(totalFrameSize)

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)
	final = append(final, encodingByte)
	final = append(final, data...)

	return final
}
