package id3v2

import (
	"fmt"
)

type FrameBinary struct {
	_FrameBase
	binData []byte
}

func (me *FrameBinary) BinData() *[]byte { return &me.binData }

func (me *FrameBinary) new(base _FrameBase, binData []byte) {
	me._FrameBase = base
	me.binData = binData
}

func (me *FrameBinary) parse(base _FrameBase, src []byte) {
	theData := make([]byte, len(src))
	copy(theData, src) // simply store bytes

	me.new(base, theData)
}

func (me *FrameBinary) Serialize() ([]byte, error) {
	totalFrameSize := 10 + len(me.binData) // header + data
	header, err := me._FrameBase.serializeHeader(totalFrameSize)
	if err != nil {
		return nil, fmt.Errorf("serializing FrameBinary header: %w", err)
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)     // 10-byte header
	final = append(final, me.binData...) // binary data as-is

	return final, nil
}
