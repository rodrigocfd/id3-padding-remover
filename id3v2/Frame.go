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
	OriginalTagSize() int // Includes 10-byte tag header.
	Flags() [2]byte
	Serialize() ([]byte, error)
}

// Constructor.
func _FrameParse(src []byte) (Frame, error) {
	header := _FrameHeaderParse(src)
	src = src[10:header.OriginalTagSize()] // skip frame header, truncate to frame size

	if header.Name4() == "COMM" {
		return _FrameCommentParse(header, src)

	} else if header.Name4()[0] == 'T' {
		texts, err := util.ParseAnyStrings(src)
		if err != nil {
			return nil, err
		}

		if len(texts) == 0 {
			return nil, fmt.Errorf("Frame %s contains no texts", header.Name4())
		} else if len(texts) == 1 {
			return _FrameTextNew(header, texts[0]), nil
		} else {
			return _FrameMultiTextNew(header, texts)
		}

	} else {
		// Anything else is treated as raw binary.
		return _FrameBinaryParse(header, src), nil
	}
}

//------------------------------------------------------------------------------

type _FrameHeader struct {
	name4        string
	originalSize int
	flags        [2]byte
}

func (me *_FrameHeader) Name4() string        { return me.name4 }
func (me *_FrameHeader) OriginalTagSize() int { return me.originalSize }
func (me *_FrameHeader) Flags() [2]byte       { return me.flags }

// Constructor.
func _FrameHeaderMake(name4 string) _FrameHeader {
	return _FrameHeader{
		name4:        name4,
		originalSize: 0,
		flags:        [2]byte{0x00, 0x00},
	}
}

// Constructor.
func _FrameHeaderParse(src []byte) _FrameHeader {
	return _FrameHeader{
		name4:        string(src[0:4]),
		originalSize: int(binary.BigEndian.Uint32(src[4:8]) + 10), // also count 10-byte tag header
		flags:        [2]byte{src[8], src[9]},
	}
}

func (me *_FrameHeader) serialize(totalFrameSize int) ([]byte, error) {
	if len(me.name4) != 4 {
		return nil, fmt.Errorf("frame name length is not 4 [%s]", me.name4)
	}

	blob := make([]byte, 0, 10) // header is 10 bytes
	blob = append(blob, []byte(me.name4)...)

	blob = util.Append32(blob, binary.BigEndian, uint32(totalFrameSize-10)) // without 10-byte header
	blob = append(blob, me.flags[0])
	blob = append(blob, me.flags[1])

	return blob, nil
}
