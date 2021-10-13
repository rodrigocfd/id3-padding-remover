package id3v2

import (
	"fmt"
	"id3fit/id3v2/util"
	"strings"
)

type FrameMultiText struct {
	_FrameBase
	texts []string
}

func (me *FrameMultiText) Texts() *[]string { return &me.texts }

func (me *FrameMultiText) parse(base _FrameBase, texts []string) error {
	if len(texts) < 2 {
		return fmt.Errorf("bad multi-text frame with only 1 text")
	}

	me._FrameBase = base
	me.texts = texts
	return nil
}

func (me *FrameMultiText) Serialize() ([]byte, error) {
	encodingByte, data := util.SerializeStrings(me.texts)
	totalFrameSize := 10 + 1 + len(data) // header + encodingByte

	header, err := me._FrameBase.serializeHeader(totalFrameSize)
	if err != nil {
		return nil, fmt.Errorf("serializing FrameMultiText header: %w", err)
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...) // 10-byte header
	final = append(final, encodingByte)
	final = append(final, data...)

	return final, nil
}

func (me *FrameMultiText) IsReplayGain() bool {
	return me._FrameBase.Name4() == "TXXX" &&
		len(me.texts) == 2 &&
		(strings.HasPrefix(me.texts[0], "replaygain_") ||
			strings.HasPrefix(me.texts[1], "replaygain_"))
}
