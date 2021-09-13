package id3

import (
	"id3fit/id3/util"
)

type FrameText struct {
	_FrameBase
	text string
}

func (me *FrameText) Text() *string { return &me.text }

func (me *FrameText) parse(base _FrameBase, texts []string) {
	me._FrameBase = base
	me.text = texts[0]
}

func (me *FrameText) Serialize() ([]byte, error) {
	encodingByte, data := util.SerializeStrings([]string{me.text})
	totalFrameSize := 10 + 1 + len(data) // header + encodingByte

	header, err := me._FrameBase.serializeHeader(totalFrameSize)
	if err != nil {
		return nil, err
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)
	final = append(final, encodingByte)
	final = append(final, data...)

	return final, nil
}
