package id3v2

import (
	"encoding/binary"
	"fmt"
	"id3fit/id3v2/util"
	"strings"
)

// A unit of data within a tag.
type Frame struct {
	name4        string    // Uniquely identifies the frame type.
	originalSize int       // Includes 10-byte frame header.
	flags        [2]byte   // Almost always zero.
	data         FrameData // Polymorphic data.
}

func (f *Frame) Name4() string     { return f.name4 }
func (f *Frame) OriginalSize() int { return f.originalSize }
func (f *Frame) Flags() [2]byte    { return f.flags }
func (f *Frame) Data() FrameData   { return f.data }

// Constructor.
func _NewFrameEmpty(name4 string) *Frame {
	return &Frame{
		name4: name4,
	}
}

// Constructor.
func _NewFrameParse(src []byte) (*Frame, error) {
	// Parse the 10-byte frame header.
	f := &Frame{
		name4:        string(src[0:4]),
		originalSize: int(binary.BigEndian.Uint32(src[4:8]) + 10), // also count 10-byte tag header
		flags:        [2]byte{src[8], src[9]},
	}

	src = src[10:f.originalSize] // skip frame header, truncate to frame size

	// Parse the frame contents.
	if f.name4 == "COMM" {
		data, err := _NewFrameDataComment(src)
		if err != nil {
			return nil, fmt.Errorf("parsing COMM: %w", err)
		}
		f.data = data

	} else if f.name4 == "APIC" {
		data, err := _NewFrameDataPicture(src)
		if err != nil {
			return nil, fmt.Errorf("parsing APIC: %w", err)
		}
		f.data = data

	} else if f.name4[0] == 'T' {
		texts, err := util.ParseAnyStrings(src)
		if err != nil {
			return nil, err
		}

		switch len(texts) {
		case 0:
			return nil, fmt.Errorf("frame %s contains no texts", f.name4)
		case 1:
			f.data = &FrameDataText{Text: texts[0]}
		case 2:
			f.data = &FrameDataUserText{Descr: texts[0], Text: texts[1]}
		default:
			return nil, fmt.Errorf("frame %s contains %d texts", f.name4, len(texts))
		}

	} else { // anything else is treated as raw binary
		f.data = _NewFrameDataBinary(src)
	}

	return f, nil
}

func (f *Frame) Serialize() []byte {
	serializedData := f.data.Serialize()

	buf := make([]byte, 0, 10+len(serializedData))
	buf = append(buf, []byte(f.name4)...)
	buf = util.Append32(buf, binary.BigEndian, uint32(len(serializedData))) // won't count 10-byte header
	buf = append(buf, f.flags[:]...)
	buf = append(buf, serializedData...)
	return buf
}

func (f *Frame) IsReplayGain() bool {
	if f.name4 == "TXXX" {
		if frameUserTxt, ok := f.data.(*FrameDataUserText); ok {
			return strings.HasPrefix(frameUserTxt.Descr, "replaygain_")
		}
	}
	return false
}
