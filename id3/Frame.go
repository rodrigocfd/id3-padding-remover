package id3

import (
	"encoding/binary"
	"fmt"
	"id3fit/id3/util"
)

// Frame is polymorphic: the underlying type will expose the methods to access
// the contents.
type Frame interface {
	Name4() string
	OriginalSize() int
	Serialize() []byte
}

// Constructor.
func _ParseFrame(src []byte) (Frame, error) {
	frameBase := _FrameBase{}
	frameBase.parse(src)
	src = src[10:frameBase.OriginalSize()] // skip frame header, truncate to frame size

	if frameBase.Name4() == "COMM" {
		frameComment := &FrameComment{}
		err := frameComment.parse(frameBase, src)
		return frameComment, err

	} else if frameBase.Name4()[0] == 'T' {
		texts, err := util.ParseAnyStrings(src)
		if err != nil {
			return nil, err
		}

		if len(texts) == 0 {
			return nil, fmt.Errorf("Frame %s contains no texts.", frameBase.Name4())

		} else if len(texts) == 1 {
			frameText := &FrameText{}
			frameText.parse(frameBase, texts)
			return frameText, nil

		} else {
			frameMultiText := &FrameMultiText{}
			err := frameMultiText.parse(frameBase, texts)
			return frameMultiText, err
		}

	} else {
		// Anything else is treated as raw binary.
		frameBinary := &FrameBinary{}
		frameBinary.parse(frameBase, src)
		return frameBinary, nil
	}
}

//------------------------------------------------------------------------------

type _FrameBase struct {
	name4        string
	originalSize int
}

func (me *_FrameBase) parse(src []byte) {
	me.name4 = string(src[0:4])
	me.originalSize = int(binary.BigEndian.Uint32(src[4:8]) + 10) // also count 10-byte tag header
}

func (me *_FrameBase) Name4() string     { return me.name4 }
func (me *_FrameBase) OriginalSize() int { return me.originalSize }

func (me *_FrameBase) serializeHeader(totalFrameSize int) []byte {
	blob := make([]byte, 0, 10) // header is 10 bytes
	blob = append(blob, []byte(me.name4)...)

	blob = util.Append32(blob, binary.BigEndian, uint32(totalFrameSize-10)) // without 10-byte header

	blob = util.Append16(blob, binary.BigEndian, 0x0000) // flags
	return blob
}
