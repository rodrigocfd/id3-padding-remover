package id3

import (
	"errors"
	"id3fit/id3/util"
	"strings"
)

type FrameMultiText struct {
	_FrameBase
	texts []string
}

func (me *FrameMultiText) parse(base *_FrameBase, texts []string) error {
	if len(texts) < 2 {
		return errors.New("Bad multi-text frame with only 1 text.")
	}

	me._FrameBase = *base
	me.texts = texts
	return nil
}

func (me *FrameMultiText) Texts() *[]string { return &me.texts }

func (me *FrameMultiText) Serialize() []byte {
	encodingByte, data := util.SerializeStrings(me.texts)
	totalFrameSize := 10 + 1 + len(data) // header + encodingByte

	header := me._FrameBase.serializeHeader(totalFrameSize)

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)
	final = append(final, encodingByte)
	final = append(final, data...)

	return final
}

func (me *FrameMultiText) IsReplayGain() bool {
	return me._FrameBase.Name4() == "TXXX" &&
		len(me.texts) == 2 &&
		(strings.HasPrefix(me.texts[0], "replaygain_") ||
			strings.HasPrefix(me.texts[1], "replaygain_"))
}
