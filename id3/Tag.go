package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Tag struct {
	version     [3]uint16
	tagSize     uint32
	paddingSize uint32
	frames      []Frame
}

func (me *Tag) Version() [3]uint16 { return me.version }
func (me *Tag) TotalSize() uint32  { return me.tagSize }
func (me *Tag) Frames() []Frame    { return me.frames }

func (me *Tag) Read(mp3Blob []byte) error {
	if !bytes.Equal(mp3Blob[:3], []byte("ID3")) {
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

	return me.readFrames(mp3Blob[10:me.tagSize]) // skip 10-byte tag header
}

func (me *Tag) readFrames(src []byte) error {
	off := 0
	for {
		if len(me.frames) == 7 {
			println("here", len(src[off:]))
		}

		if me.isSliceZeroed(src[off:]) { // we entered a padding region after all frames
			me.paddingSize = uint32(len(src[off:]))
			break
		} else if off == int(me.tagSize) { // end of tag, no padding
			break
		}

		me.frames = append(me.frames, Frame{})
		lastFrame := &me.frames[len(me.frames)-1]

		err := lastFrame.Read(src[off:])
		if err != nil {
			return err
		}
		off += int(lastFrame.frameSize) // now points to 1st byte of next frame
	}

	return nil
}

func (me *Tag) isSliceZeroed(blob []byte) bool {
	for _, b := range blob {
		if b != 0x00 {
			return false
		}
	}
	return true
}
