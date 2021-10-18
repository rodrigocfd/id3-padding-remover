package id3v2

import (
	"encoding/binary"
	"fmt"
	"id3fit/id3v2/util"
)

// Frame is polymorphic: the underlying type will expose the methods to access
// the contents.
type Frame interface {
	Name4() string
	OriginalSize() int
	Serialize() ([]byte, error)
}

// Constructor.
func _ParseFrame(src []byte) (Frame, error) {
	frameBase := _ParseFrameBase(src)
	src = src[10:frameBase.OriginalSize()] // skip frame header, truncate to frame size

	if frameBase.Name4() == "COMM" {
		return _ParseFrameComment(frameBase, src)

	} else if frameBase.Name4()[0] == 'T' {
		texts, err := util.ParseAnyStrings(src)
		if err != nil {
			return nil, err
		}

		if len(texts) == 0 {
			return nil, fmt.Errorf("Frame %s contains no texts", frameBase.Name4())
		} else if len(texts) == 1 {
			return _NewFrameText(frameBase, texts[0]), nil
		} else {
			return _NewFrameMultiText(frameBase, texts)
		}

	} else {
		// Anything else is treated as raw binary.
		return _ParseFrameBinary(frameBase, src), nil
	}
}

//------------------------------------------------------------------------------

type _FrameBase struct {
	name4        string
	originalSize int
}

// Constructor.
func _MakeFrameBase(name4 string) _FrameBase {
	return _FrameBase{
		name4:        name4,
		originalSize: 0,
	}
}

// Constructor.
func _ParseFrameBase(src []byte) _FrameBase {
	return _FrameBase{
		name4:        string(src[0:4]),
		originalSize: int(binary.BigEndian.Uint32(src[4:8]) + 10), // also count 10-byte tag header
	}
}

func (me *_FrameBase) Name4() string     { return me.name4 }
func (me *_FrameBase) OriginalSize() int { return me.originalSize }

func (me *_FrameBase) serializeHeader(totalFrameSize int) ([]byte, error) {
	if len(me.name4) != 4 {
		return nil, fmt.Errorf("frame name length is not 4 [%s]", me.name4)
	}

	blob := make([]byte, 0, 10) // header is 10 bytes
	blob = append(blob, []byte(me.name4)...)

	blob = util.Append32(blob, binary.BigEndian, uint32(totalFrameSize-10)) // without 10-byte header

	blob = util.Append16(blob, binary.BigEndian, 0x0000) // flags
	return blob, nil
}
