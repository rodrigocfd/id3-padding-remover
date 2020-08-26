package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"wingows/gui"
)

type Tag struct {
	totalTagSize uint32
	paddingSize  uint32
	frames       []Frame
}

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

	return me.ReadFromBinary(fMap.HotSlice())
}

func (me *Tag) ReadFromBinary(src []byte) error {
	me.totalTagSize = 0 // clear
	me.paddingSize = 0
	me.frames = nil

	if err := me.parseTagHeader(src); err != nil {
		return err
	}

	src = src[10:me.totalTagSize] // skip 10-byte tag header; truncate to tag size
	if err := me.parseAllFrames(src); err != nil {
		return err
	}

	return nil // tag parsed successfully
}

func (me *Tag) SerializeToFile(mp3Path string) error {
	// Serialize all frames.
	serializedFrames := make([][]byte, len(me.frames))
	tagSize := 0
	for i := range me.frames {
		serialized := _FrameSerializer.SerializeFrame(me.frames[i])
		serializedFrames[i] = serialized
		tagSize += len(serialized)
	}

	// Build the binary blob.
	blob := make([]byte, 10, 10+tagSize)
	copy(blob, []byte("ID3"))    // magic bytes
	copy(blob[3:], []byte{3, 0}) // v2.3.0

	blob[5] = 0 // flags
	binary.BigEndian.PutUint32(blob[6:], _Util.SynchSafeEncode(uint32(tagSize)))

	for _, serialized := range serializedFrames {
		blob = append(blob, serialized...)
	}

	return me.writeTagToFile(mp3Path, blob)
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
	if !bytes.Equal(src[3:5], []byte{3, 0}) { // the first "2" is not stored in the tag
		return errors.New(
			fmt.Sprintf("Tag version 2.%d.%d is not supported, only 2.3.0.",
				src[1], src[2]))
	}

	// Validate unsupported flags.
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

		newFrame, err := _FrameParser.ParseFrame(src)
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

func (me *Tag) writeTagToFile(mp3Path string, blob []byte) error {
	fout := gui.FileMapped{}
	if err := fout.OpenExistingForReadWrite(mp3Path); err != nil {
		return err
	}
	defer fout.Close()
	fileMem := fout.HotSlice()

	currentTag := Tag{}
	currentTag.ReadFromBinary(fileMem)

	diff := len(blob) - int(currentTag.TotalTagSize()) // size difference between new/old tags

	if diff > 0 { // new tag is larger, we need to make room
		if err := fout.SetSize(fout.Size() + uint64(diff)); err != nil {
			return err
		}
	}

	// Move the MP3 data block inside the file.
	copy(fileMem[int(currentTag.TotalTagSize())+diff:], fileMem[currentTag.TotalTagSize():])

	// Copy the new tag into the file, no padding.
	copy(fileMem, blob)

	if diff < 0 { // new tag is shorter
		if err := fout.SetSize(fout.Size() + uint64(diff)); err != nil {
			return err
		}
	}

	return nil
}
