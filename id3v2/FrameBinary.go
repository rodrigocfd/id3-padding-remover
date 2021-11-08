package id3v2

import (
	"fmt"
)

type FrameBinary struct {
	_FrameHeader
	binData []byte
}

// Constructor.
func _FrameBinaryNew(header _FrameHeader, binData []byte) *FrameBinary {
	return &FrameBinary{
		_FrameHeader: header,
		binData:      binData,
	}
}

// Constructor.
func _FrameBinaryParse(base _FrameHeader, src []byte) *FrameBinary {
	theData := make([]byte, len(src))
	copy(theData, src) // simply store bytes

	return _FrameBinaryNew(base, theData)
}

func (me *FrameBinary) BinData() *[]byte { return &me.binData }

func (me *FrameBinary) Serialize() ([]byte, error) {
	totalFrameSize := 10 + len(me.binData) // headerBlob + data
	headerBlob, err := me._FrameHeader.serialize(totalFrameSize)
	if err != nil {
		return nil, fmt.Errorf("serializing FrameBinary header: %w", err)
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, headerBlob...) // 10-byte header
	final = append(final, me.binData...) // binary data as-is

	return final, nil
}
