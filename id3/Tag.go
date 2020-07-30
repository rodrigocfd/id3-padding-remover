package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"wingows/gui"
)

type Tag struct {
	version     [3]uint16
	tagSize     uint32
	paddingSize uint32
	frames      []Frame
}

func (me *Tag) Version() [3]uint16  { return me.version }
func (me *Tag) TotalSize() uint32   { return me.tagSize }
func (me *Tag) PaddingSize() uint32 { return me.paddingSize }
func (me *Tag) Frames() []Frame     { return me.frames }

func (me *Tag) Album() (string, bool)  { return me.simpleText("TALB") }
func (me *Tag) Artist() (string, bool) { return me.simpleText("TPE1") }
func (me *Tag) Genre() (string, bool)  { return me.simpleText("TCON") }
func (me *Tag) Title() (string, bool)  { return me.simpleText("TIT2") }
func (me *Tag) Track() (string, bool)  { return me.simpleText("TRCK") }
func (me *Tag) Year() (string, bool)   { return me.simpleText("TYER") }

func (me *Tag) Picture() ([]byte, bool) {
	frame := me.findFrame("TALB")
	if frame == nil {
		return nil, false // frame not found
	}
	return frame.binData, true // return whole binary data
}

func (me *Tag) Comment() ([]string, bool) {
	frame := me.findFrame("COMM")
	if frame == nil {
		return nil, false // frame not found
	}
	return frame.texts, true // return all comment strings
}

func (me *Tag) ReadBinary(mp3Blob []byte) error {
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
		binary.BigEndian.Uint32(mp3Blob[6:10]), // also count 10-byte tag header
	) + 10

	return me.readFrames(mp3Blob[10:me.tagSize]) // skip 10-byte tag header
}

func (me *Tag) ReadFile(path string) error {
	fMap := gui.FileMapped{}
	fMap.OpenExistingForRead(path)
	defer fMap.Close() // HotSlice() needs the file to remain open

	contents := fMap.HotSlice()
	return me.ReadBinary(contents)
}

func (me *Tag) readFrames(src []byte) error {
	for {
		if len(src) == 0 { // end of tag, no padding found
			break
		} else if isSliceZeroed(src) { // we entered a padding region after all frames
			me.paddingSize = uint32(len(src)) // store padding size
			break
		}

		me.frames = append(me.frames, Frame{}) // append new frame to our slice
		newFrame := &me.frames[len(me.frames)-1]
		err := newFrame.Read(src) // parse frame contents
		if err != nil {
			return err // an error occurred when parsing the frame
		}

		if int(newFrame.frameSize) > len(src) {
			return errors.New("Frame size is greater than real size.")
		}

		src = src[newFrame.frameSize:] // now starts at 1st byte of next frame
	}
	return nil // all frames parsed
}

func (me *Tag) findFrame(name4 string) *Frame {
	for i := range me.frames {
		if me.frames[i].name4 == name4 {
			return &me.frames[i]
		}
	}
	return nil
}

func (me *Tag) simpleText(name4 string) (string, bool) {
	frame := me.findFrame(name4)
	if frame == nil {
		return "", false // frame not found
	} else if frame.kind != FRAME_KIND_TEXT {
		return "", false // not a text frame
	}
	return frame.texts[0], true // return 1st text of frame
}
