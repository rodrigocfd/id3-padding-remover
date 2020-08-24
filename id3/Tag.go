package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"wingows/gui"
)

type Tag struct {
	version      [3]uint16
	totalTagSize uint32
	paddingSize  uint32
	frames       []Frame
}

func (me *Tag) Version() [3]uint16   { return me.version }
func (me *Tag) TotalTagSize() uint32 { return me.totalTagSize }
func (me *Tag) PaddingSize() uint32  { return me.paddingSize }
func (me *Tag) Frames() []Frame      { return me.frames }

func (me *Tag) Album() *FrameText    { return me.findByName4("TALB").(*FrameText) }
func (me *Tag) Artist() *FrameText   { return me.findByName4("TPE1").(*FrameText) }
func (me *Tag) Composer() *FrameText { return me.findByName4("TCOM").(*FrameText) }
func (me *Tag) Genre() *FrameText    { return me.findByName4("TCON").(*FrameText) }
func (me *Tag) Title() *FrameText    { return me.findByName4("TIT2").(*FrameText) }
func (me *Tag) Track() *FrameText    { return me.findByName4("TRCK").(*FrameText) }
func (me *Tag) Year() *FrameText     { return me.findByName4("TYER").(*FrameText) }

func (me *Tag) ReadFromFile(mp3Path string) error {
	fMap := gui.FileMapped{}
	fMap.OpenExistingForRead(mp3Path)
	defer fMap.Close() // HotSlice() needs the file to remain open

	src := fMap.HotSlice()

	if err := me.parseTagHeader(src); err != nil {
		return err
	}

	src = src[10:me.totalTagSize] // skip 10-byte tag header; truncate to tag size
	if err := me.parseAllFrames(src); err != nil {
		return err
	}

	return nil // tag parsed successfully
}

func (me *Tag) findByName4(name4 string) Frame {
	for _, myFrame := range me.Frames() {
		if myFrame.Name4() == name4 {
			return myFrame
		}
	}
	return nil // not found
}

func (me *Tag) parseTagHeader(src []byte) error {
	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return errors.New("No ID3 tag found.")
	}

	// Validate tag version 2.3.0.
	me.version = [3]uint16{
		2, // the "2" is not actually stored in the tag itself
		uint16(src[3]),
		uint16(src[4]),
	}
	if me.version[1] != 3 && me.version[2] != 0 { // not v2.3.0?
		return errors.New(
			fmt.Sprintf("Tag version 2.%d.%d is not supported, only 2.3.0.",
				me.version[1], me.version[2]),
		)
	}

	// Validade unsupported flags.
	if (src[5] & 0b1000_0000) != 0 {
		return errors.New("Tag is unsynchronised, not supported.")
	} else if (src[5] & 0b0100_0000) != 0 {
		return errors.New("Tag extended header not supported.")
	}

	// Read tag size.
	me.totalTagSize = _Util.SynchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]), // also count 10-byte tag header
	) + 10

	return nil
}

func (me *Tag) parseAllFrames(src []byte) error {
	for {
		if len(src) == 0 { // end of tag, no padding found
			break
		} else if _Util.IsSliceZeroed(src) { // we entered a padding region after all frames
			me.paddingSize = uint32(len(src)) // store padding size
			break
		}

		newFrame, err := _ParseFrame(src)
		if err != nil {
			return err // error when parsing the frame
		}
		me.frames = append(me.frames, newFrame) // add the frame to our collection

		if newFrame.TotalFrameSize() > uint(len(src)) { // means the tag was serialized with error
			return errors.New("Frame size is greater than real size.")
		}

		src = src[newFrame.TotalFrameSize():] // now starts at 1st byte of next frame
	}

	return nil // all frames parsed successfully
}
