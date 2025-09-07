//go:build windows

package id3v2

import (
	"encoding/binary"

	"github.com/rodrigocfd/xslices"
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

// Serializes the frame into bytes.
func (me *Frame) Serialize() []byte {
	body := me.body.Serialize()

	szBlob := 10 + len(body) // 10-byte header + body
	blob := make([]byte, 0, szBlob)
	blob = append(blob, []byte(me.name4)...)

	blob = xslices.AppendN(blob, 4, 0x00)
	binary.BigEndian.PutUint32(blob[4:8], uint32(len(body))) // don't count 10-byte header size

	blob = append(blob, me.flags[:]...)
	blob = append(blob, body...)

	return blob
}
