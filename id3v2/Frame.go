//go:build windows

package id3v2

import (
	"encoding/binary"
	"fmt"

	"github.com/rodrigocfd/windigo/win"
)

type Frame struct {
	name4        string
	declaredSize uint // Used only at parsing.
	flags        [2]byte
	body         Body // Polymorphic.
}

func (me *Frame) Name4() string      { return me.name4 }
func (me *Frame) DeclaredSize() uint { return me.declaredSize }
func (me *Frame) Body() Body         { return me.body }

// Constructor.
func _FrameNewWithText(name4, text string) *Frame {
	me := &Frame{
		name4: name4,
	}

	if name4 == "COMM" {
		me.body = &BodyComment{
			Lang3: "eng",
			Text:  text,
		}
	} else {
		me.body = &BodyText{ // assume simple text frame
			Text: text,
		}
	}
	return me
}

// Constructor.
func _FrameParse(src []byte) (*Frame, error) {
	// Parse the 10-byte frame header.
	me := &Frame{
		name4:        string(src[0:4]),
		declaredSize: uint(binary.BigEndian.Uint32(src[4:8]) + 10), // also count 10-byte tag header
		flags:        [2]byte{src[8], src[9]},
	}
	if me.declaredSize > uint(len(src)) {
		me.declaredSize = uint(len(src)) // if serialized with error, be complacent
	}

	src = src[10:me.declaredSize] // skip frame header, truncate to declared frame size

	if err := me.parseBody(src); err != nil {
		return nil, err
	}
	return me, nil
}

func (me *Frame) parseBody(src []byte) error {
	if me.name4 == "COMM" {
		comm, err := _BodyCommentParse(src)
		if err != nil {
			return err
		}
		me.body = comm
	} else if me.name4 == "APIC" {
		apic, err := _BodyPictureParse(src)
		if err != nil {
			return err
		}
		me.body = apic
	} else if me.name4[0] == 'T' {
		texts, err := parseStrings(src)
		if err != nil {
			return fmt.Errorf("frame %s with bad strings: %w", me.name4, err)
		}

		switch len(texts) {
		case 0:
			return fmt.Errorf("frame %s contains no texts", me.name4)
		case 1:
			me.body = &BodyText{Text: texts[0]}
		case 2:
			me.body = &BodyUserText{Descr: texts[0], Text: texts[1]}
		default:
			return fmt.Errorf("frame %s contains %d texts", me.name4, len(texts))
		}
	} else { // everything else is treated as raw binary
		me.body = _BodyBinaryParse(src)
	}
	return nil
}

// Returns a newly allocated frame with a copy of all the data.
func (me *Frame) Clone() *Frame {
	return &Frame{me.name4, me.declaredSize, me.flags, me.body.Clone()}
}

// Serializes the frame into bytes.
func (me *Frame) Serialize(pDest *win.Vec[byte]) uint {
	pDest.Reserve(pDest.Len() + 10) // header size
	pDest.Append([]byte(me.name4)...)

	bodySizeOffset := pDest.Len()
	pDest.AppendN(4, 0x00) // placeholder for body size
	pDest.Append(me.flags[:]...)

	szBody := me.body.Serialize(pDest) // won't count 10-byte frame header
	binary.BigEndian.PutUint32(pDest.HotSlice()[bodySizeOffset:], uint32(szBody))
	return 10 + szBody // count 10-byte frame header
}
