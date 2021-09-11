package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"id3fit/id3/util"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

// Metadata of a single MP3 file, composed of many frames.
type Tag struct {
	originalSize    int
	originalPadding int
	frames          []Frame
}

func (me *Tag) OriginalSize() int    { return me.originalSize }
func (me *Tag) OriginalPadding() int { return me.originalPadding }
func (me *Tag) Frames() []Frame      { return me.frames }

// Public constructor; reads the tag from an MP3 file.
func ReadTagFromFile(mp3Path string) (*Tag, error) {
	tag := &Tag{}
	return tag, tag.readFromFile(mp3Path)
}

// Public constructor; reads the tag from a binary blob.
func ReadTagFromBinary(src []byte) (*Tag, error) {
	tag := &Tag{}
	return tag, tag.readFromBinary(src)
}

func (me *Tag) readFromFile(mp3Path string) error {
	fMap, err := win.OpenFileMapped(mp3Path, co.OPEN_FILEMAP_MODE_READ)
	if err != nil {
		return err
	}
	defer fMap.Close()

	return me.readFromBinary(fMap.HotSlice())
}

func (me *Tag) readFromBinary(src []byte) error {
	originalSize, err := me.parseTagHeader(src)
	if err != nil {
		return err
	}
	src = src[10:originalSize] // skip 10-byte tag header; truncate to tag bounds

	frames, originalPadding, err := me.parseAllFrames(src)
	if err != nil {
		return err
	}

	me.originalSize = originalSize
	me.originalPadding = originalPadding
	me.frames = frames
	return nil
}

func (me *Tag) parseTagHeader(src []byte) (int, error) {
	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return 0, errors.New("No ID3 tag found.")
	}

	// Validate tag version 2.3.0.
	if !bytes.Equal(src[3:5], []byte{3, 0}) { // the first "2" is not stored in the tag
		return 0, errors.New(
			fmt.Sprintf("Tag version 2.%d.%d is not supported, only 2.3.0.",
				src[3], src[4]),
		)
	}

	// Validate unsupported flags.
	if (src[5] & 0b1000_0000) != 0 {
		return 0, errors.New("Tag is unsynchronised, not supported.")
	} else if (src[5] & 0b0100_0000) != 0 {
		return 0, errors.New("Tag extended header not supported.")
	}

	// Read tag size.
	originalSize := int(util.SynchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]), // also count 10-byte tag header
	) + 10)

	return originalSize, nil
}

func (me *Tag) parseAllFrames(src []byte) ([]Frame, int, error) {
	frames := make([]Frame, 0, 6) // arbitrary capacity
	padding := 0

	for {
		if len(src) == 0 { // end of tag, no padding found
			break
		} else if util.IsSliceZeroed(src) { // we entered a padding region after all frames
			padding = len(src) // store padding size
			break
		}

		newFrame, err := _ParseFrame(src)
		if err != nil {
			return nil, 0, err // error when parsing the frame
		}
		if newFrame.OriginalSize() > len(src) { // means the tag was serialized with error
			return nil, 0, errors.New("Frame size is greater than real size.")
		}
		frames = append(frames, newFrame) // add the frame to our collection

		src = src[newFrame.OriginalSize():] // now starts at 1st byte of next frame
	}

	return frames, padding, nil
}

func (me *Tag) Serialize() []byte {
	data := make([]byte, 0, 100) // arbitrary; all serialized frames
	for _, frame := range me.frames {
		data = append(data, frame.Serialize()...)
	}

	final := make([]byte, 0, 10+len(data))       // header
	final = append(final, []byte("ID3")...)      // magic bytes
	final = append(final, []byte{0x03, 0x00}...) // tag version
	final = append(final, 0x00)                  // flags

	synchSafeDataSize := util.SynchSafeEncode(uint32(len(data)))
	final = util.Append32(final, binary.BigEndian, synchSafeDataSize)

	final = append(final, data...)
	return final
}

func (me *Tag) SerializeToFile(mp3Path string) error {
	newTag := me.Serialize()

	fout, err := win.OpenFileMapped(mp3Path, co.OPEN_FILEMAP_MODE_RW)
	if err != nil {
		return err
	}
	defer fout.Close()
	fileMem := fout.HotSlice()

	currentTag, err := ReadTagFromBinary(fileMem)
	if err != nil {
		return err
	}

	diff := len(newTag) - currentTag.OriginalSize() // size difference between new/old tags

	if diff > 0 { // new tag is larger, we need to make room
		if err := fout.Resize(fout.Size() + diff); err != nil {
			return err
		}
	}

	// Move the MP3 data block inside the file, back or forth.
	copy(fileMem[int(currentTag.OriginalSize())+diff:], fileMem[currentTag.OriginalSize():])

	// Copy the new tag into the file, no padding.
	copy(fileMem, newTag)

	if diff < 0 { // new tag is shorter, shrink
		if err := fout.Resize(fout.Size() + diff); err != nil {
			return err
		}
	}

	return nil
}

func (me *Tag) DeleteFrames(fun func(f Frame) bool) {
	newSlice := make([]Frame, 0, len(me.frames))

	for _, f := range me.frames {
		willDelete := fun(f)
		if !willDelete { // the new slice will contain the non-deleted tags
			newSlice = append(newSlice, f)
		}
	}

	me.frames = newSlice // throw the old one away
}

func (me *Tag) FrameByName(name4 string) (Frame, bool) {
	for _, f := range me.frames {
		if f.Name4() == name4 {
			return f, true
		}
	}
	return nil, false
}
