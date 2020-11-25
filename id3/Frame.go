package id3

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Frame is polymorphic, the underlying type will expose the methods to access the contents.
// Note that changing name4 is not allowed.
type Frame interface {
	Name4() string
	Serialize() []byte
}

type _FrameBase struct { // implements Frame
	name4 string
}

func (me *_FrameBase) Name4() string { return me.name4 }

// Constructor.
func _ParseFrame(src []byte) (Frame, int, error) {
	fr := _FrameBase{
		name4: string(src[0:4]),
	}

	totalFrameSize := int(binary.BigEndian.Uint32(src[4:8]) + 10) // also count 10-byte tag header

	src = src[10:totalFrameSize] // skip frame header, truncate to frame size

	if fr.name4 == "COMM" {
		frComm, err := _ParseFrameComment(&fr, src)
		if err != nil {
			return nil, 0, err
		}
		return frComm, totalFrameSize, nil

	} else if fr.name4[0] == 'T' {
		texts, err := _Util.ParseAnyStrings(src)
		if err != nil {
			return nil, 0, err
		}

		if len(texts) == 0 {
			return nil, 0, errors.New(fmt.Sprintf("Frame %s contains no texts.", fr.name4))

		} else if len(texts) == 1 {
			return _ParseFrameText(&fr, texts), totalFrameSize, nil

		} else {
			return _ParseFrameMultiText(&fr, texts), totalFrameSize, nil
		}
	}

	// Anything else is treated as raw binary.
	return _ParseFrameBinary(&fr, src), totalFrameSize, nil
}

func (me *_FrameBase) serializeHeader(totalFrameSize int) []byte {
	blob := make([]byte, 0, 10) // header is 10 bytes
	blob = append(blob, []byte(me.name4)...)

	blob = append(blob, []byte{0, 0, 0, 0}...)
	binary.BigEndian.PutUint32(blob[4:8], uint32(totalFrameSize-10)) // without 10-byte header

	blob = append(blob, []byte{0, 0}...) // flags
	return blob
}
