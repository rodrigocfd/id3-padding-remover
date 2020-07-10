package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Tag struct {
	version [3]uint16
	tagSize uint32
	frames  []Frame
}

func (me *Tag) Version() [3]uint16 { return me.version }
func (me *Tag) TotalSize() uint32  { return me.tagSize }
func (me *Tag) Frames() []Frame    { return me.frames }

func (me *Tag) Read(mp3Blob []uint8) error {
	if !bytes.Equal(mp3Blob[:3], []uint8("ID3")) {
		return errors.New("No tag found.")
	}

	me.version = [3]uint16{
		2, // the "2" is not actually stored in the tag itself
		uint16(mp3Blob[3]),
		uint16(mp3Blob[4]),
	}
	if me.version[1] != 3 && me.version[2] != 0 { // not v2.3.0?
		return errors.New(
			fmt.Sprintf("Tag version 2.%d.%d is not supported, only 2.3.0.",
				me.version[1], me.version[2]),
		)
	}

	if (mp3Blob[5] & 0b1000_0000) != 0 { // flags
		return errors.New("Tag is unsynchronised, not supported.")
	} else if (mp3Blob[5] & 0b0100_0000) != 0 {
		return errors.New("Tag extended header not supported.")
	}

	me.tagSize = synchSafeDecode(
		binary.BigEndian.Uint32(mp3Blob[6:10]) + 10, // also count 10-byte tag header
	)

	return me.readFrames(mp3Blob)
}

func (me *Tag) readFrames(mp3Blob []uint8) error {
	off := uint32(10) // skip 10-byte tag header

	for {
		// CHECK IF PADDING BEFORE PARSING AS A FRAME

		me.frames = append(me.frames, Frame{})
		lastFrame := &me.frames[len(me.frames)-1]

		err := lastFrame.Read(mp3Blob[off:])
		if err != nil {
			return err
		}
		off += lastFrame.frameSize // now points to 1st byte of next frame

		if off == me.tagSize { // end of tag
			break
		}
	}

	return nil
}
