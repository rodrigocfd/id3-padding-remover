//go:build windows

package id3v2

import (
	"encoding/binary"
)

type Frame struct {
	name4        string
	declaredSize int // Used only at parsing.
	flags        [2]byte
	body         Body // Polymorphic.
}

func (me *Frame) Name4() string     { return me.name4 }
func (me *Frame) DeclaredSize() int { return me.declaredSize }
func (me *Frame) Body() Body        { return me.body }

// Constructor.
func _FrameParse(src []byte) (*Frame, error) {
	// Parse the 10-byte frame header.
	name4 := string(src[0:4])
	declaredSize := int(binary.BigEndian.Uint32(src[4:8]) + 10) // also count 10-byte tag header
	flags := [2]byte{src[8], src[9]}

	if declaredSize > len(src) {
		declaredSize = len(src) // if serialized with error, be complacent
	}

	src = src[10:declaredSize] // skip frame header, truncate to declared frame size

	// Parse the data body.
	body, err := _BodyParse(name4, src)
	if err != nil {
		return nil, err
	}

	return &Frame{name4, declaredSize, flags, body}, nil
}

// Constructor.
func _FrameNewWithText(name4, text string) *Frame {
	switch name4 {
	case "COMM":
		return &Frame{
			name4: name4,
			body: &BodyComment{
				Lang3: "eng",
				Descr: "",
				Text:  text,
			},
		}
	default: // otherwise assume simple text frame
		return &Frame{
			name4: name4,
			body: &BodyText{
				Text: text,
			},
		}
	}
}

// Returns a newly allocated frame with a copy of all the data.
func (me *Frame) Clone() *Frame {
	return &Frame{me.name4, me.declaredSize, me.flags, me.body.Clone()}
}

// The size in bytes required to serialize the frame.
func (me *Frame) SerializeSize() int {
	return 10 + me.body.SerializeSize() // start with 10-byte frame header
}

// Serializes the frame into bytes. The dest buffer will be appended.
func (me *Frame) Serialize(dest []byte) []byte {
	dest = append(dest, []byte(me.name4)...)

	szBody := me.body.SerializeSize() // don't count 10-byte header size
	dest = appendUint32(dest, binary.BigEndian, uint32(szBody))

	dest = append(dest, me.flags[:]...)
	dest = me.body.Serialize(dest)
	return dest
}
