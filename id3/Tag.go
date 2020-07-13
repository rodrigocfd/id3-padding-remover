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
	frames      map[string]Frame
}

func (me *Tag) Version() [3]uint16       { return me.version }
func (me *Tag) TotalSize() uint32        { return me.tagSize }
func (me *Tag) Frames() map[string]Frame { return me.frames }

func (me *Tag) Album() (string, bool)  { return me.simpleText("TALB") }
func (me *Tag) Artist() (string, bool) { return me.simpleText("TPE1") }
func (me *Tag) Genre() (string, bool)  { return me.simpleText("TCON") }
func (me *Tag) Title() (string, bool)  { return me.simpleText("TIT2") }
func (me *Tag) Track() (string, bool)  { return me.simpleText("TRCK") }
func (me *Tag) Year() (string, bool)   { return me.simpleText("TYER") }

func (me *Tag) AlbumArt() ([]byte, bool) {
	if frame, ok := me.frames["TALB"]; ok {
		return frame.binData, true // return whole binary data
	}
	return nil, false
}

func (me *Tag) Comment() ([]string, bool) {
	if frame, ok := me.frames["COMM"]; ok {
		return frame.texts, true // return all comment strings
	}
	return nil, false
}

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
	me.frames = make(map[string]Frame)
	off := 0

	for {
		if me.isSliceZeroed(src[off:]) { // we entered a padding region after all frames
			me.paddingSize = uint32(len(src[off:])) // store padding size
			break
		} else if off == int(me.tagSize) { // end of tag, no padding found
			break
		}

		newFrame := Frame{}
		err := newFrame.Read(src[off:])
		if err != nil {
			return err
		}
		me.frames[newFrame.name4] = newFrame // save new frame into map
		off += int(newFrame.frameSize)       // now points to 1st byte of next frame
	}

	return nil
}

func (me *Tag) isSliceZeroed(blob []byte) bool {
	for _, b := range blob {
		if b != 0x00 {
			return false
		}
	}
	return true // the slice only contain zeros
}

func (me *Tag) simpleText(name4 string) (string, bool) {
	if frame, ok := me.frames[name4]; ok {
		if frame.kind != FRAME_TEXT {
			return "", false // not a text frame
		}
		return frame.texts[0], true // returns 1st text of frame
	}
	return "", false // frame not found
}
