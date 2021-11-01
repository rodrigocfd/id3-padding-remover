package id3v2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"id3fit/id3v2/util"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

// Metadata of a single MP3 file, composed of many frames.
type Tag struct {
	originalSize    int
	originalPadding int
	frames          []Frame
}

// Constructor; creates a new tag with no frames.
func TagNewEmpty() *Tag { return &Tag{} }

// Constructor; reads the tag from an MP3 file.
func TagReadFromFile(mp3Path string) (*Tag, error) {
	me := TagNewEmpty()
	if err := me.readFromFile(mp3Path); err != nil {
		return nil, err
	}
	return me, nil
}

// Constructor; reads the tag from a binary blob.
func TagReadFromBinary(src []byte) (*Tag, error) {
	me := &Tag{}
	if err := me.readFromBinary(src); err != nil {
		return nil, err
	}
	return me, nil
}

func (me *Tag) OriginalSize() int    { return me.originalSize }
func (me *Tag) OriginalPadding() int { return me.originalPadding }
func (me *Tag) Frames() []Frame      { return me.frames }
func (me *Tag) IsEmpty() bool        { return len(me.frames) == 0 }

func (me *Tag) readFromFile(mp3Path string) error {
	fMap, err := win.FileMappedOpen(mp3Path, co.FILE_OPEN_READ_EXISTING)
	if err != nil {
		return fmt.Errorf("opening file to read: %w", err)
	}
	defer fMap.Close()

	return me.readFromBinary(fMap.HotSlice())
}

func (me *Tag) readFromBinary(src []byte) error {
	originalSize, err := me.parseTagHeader(src)
	if err != nil {
		return fmt.Errorf("parsing tag header: %w", err)
	}
	src = src[10:originalSize] // skip 10-byte tag header; truncate to tag bounds

	frames, originalPadding, err := me.parseAllFrames(src)
	if err != nil {
		return fmt.Errorf("parsing all frames: %w", err)
	}

	me.originalSize = originalSize
	me.originalPadding = originalPadding
	me.frames = frames
	return nil
}

func (me *Tag) parseTagHeader(src []byte) (tagSize int, e error) {
	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return 0, &ErrorNoTagFound{}
	}

	// Validate tag version 2.3.0.
	if !bytes.Equal(src[3:5], []byte{3, 0}) { // the first "2" is not stored in the tag
		return 0, fmt.Errorf(
			"tag version 2.%d.%d is not supported, only 2.3.0",
			src[3], src[4])
	}

	// Validate unsupported flags.
	if (src[5] & 0b1000_0000) != 0 {
		return 0, fmt.Errorf("unsynchronised tag not supported")
	} else if (src[5] & 0b0100_0000) != 0 {
		return 0, fmt.Errorf("tag extended header not supported")
	}

	// Read and validate tag size.
	mp3Off, hasMp3Off := util.FindMp3Signature(src)
	writtenTagSize := int(util.SynchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]), // also count 10-byte tag header
	) + 10)

	if hasMp3Off && mp3Off != writtenTagSize {
		return 0, fmt.Errorf("bad written tag size: %d (actual %d)", writtenTagSize, mp3Off)
	}
	return writtenTagSize, nil
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
			return nil, 0, fmt.Errorf("parsing frames: %w", err) // error when parsing the frame
		}
		if newFrame.OriginalTagSize() > len(src) { // means the tag was serialized with error
			return nil, 0, fmt.Errorf("frame size is greater than real size")
		}
		frames = append(frames, newFrame) // add the frame to our collection

		src = src[newFrame.OriginalTagSize():] // now starts at 1st byte of next frame
	}

	return frames, padding, nil
}

func (me *Tag) Serialize() ([]byte, error) {
	framesBlob := make([]byte, 0, 100) // arbitrary; all serialized frames
	for _, frame := range me.frames {
		if frameData, err := frame.Serialize(); err != nil {
			return nil, fmt.Errorf("serializing to string: %w", err)
		} else {
			framesBlob = append(framesBlob, frameData...) // append the frame bytes to the big blob
		}
	}

	final := make([]byte, 0, 10+len(framesBlob)) // header + serialized frames
	final = append(final, []byte("ID3")...)      // magic bytes
	final = append(final, []byte{0x03, 0x00}...) // tag version 2.3.0
	final = append(final, 0x00)                  // flags

	synchSafeDataSize := util.SynchSafeEncode(uint32(len(framesBlob)))
	final = util.Append32(final, binary.BigEndian, synchSafeDataSize)

	final = append(final, framesBlob...)
	return final, nil
}

func (me *Tag) SerializeToFile(mp3Path string) error {
	newTag := []byte{} // if tag is empty, this will actually remove any existing tag
	if !me.IsEmpty() {
		if theNewTag, err := me.Serialize(); err != nil {
			return fmt.Errorf("serializing tag: %w", err)
		} else {
			newTag = theNewTag
		}
	}

	fout, err := win.FileMappedOpen(mp3Path, co.FILE_OPEN_RW_EXISTING)
	if err != nil {
		return fmt.Errorf("opening file to serialize: %w", err)
	}
	defer fout.Close()
	foutMem := fout.HotSlice()

	currentTag, err := TagReadFromBinary(foutMem)
	if err != nil {
		return fmt.Errorf("reading current tag: %w", err)
	}

	diff := len(newTag) - currentTag.OriginalSize() // size difference between new/old tags

	if diff > 0 { // new tag is larger, we need to make room
		if err := fout.Resize(fout.Size() + diff); err != nil {
			return fmt.Errorf("increasing file room: %w", err)
		}
	}

	// Move the MP3 data block inside the file, back or forth.
	copy(foutMem[int(currentTag.OriginalSize())+diff:], foutMem[currentTag.OriginalSize():])

	// Copy the new tag into the file, no padding.
	copy(foutMem, newTag)

	if diff < 0 { // new tag is shorter, shrink
		if err := fout.Resize(fout.Size() + diff); err != nil {
			return fmt.Errorf("decreasing file room: %w", err)
		}
	}

	return nil
}

func (me *Tag) DeleteFrames(fun func(f Frame) (willDelete bool)) {
	newSlice := make([]Frame, 0, len(me.frames))

	for _, f := range me.frames {
		willDelete := fun(f)
		if !willDelete { // the new slice will contain the non-deleted tags
			newSlice = append(newSlice, f)
		}
	}

	me.frames = newSlice // throw the old one away
}

func (me *Tag) FrameByName4(name4 string) (Frame, bool) {
	for _, f := range me.frames {
		if f.Name4() == name4 {
			return f, true
		}
	}
	return nil, false
}

func (me *Tag) TextByName4(name4 TEXT) (string, bool) {
	if frDyn, has := me.FrameByName4(string(name4)); has {
		switch fr := frDyn.(type) {
		case *FrameText:
			return *fr.Text(), true
		case *FrameComment:
			return *fr.Text(), true
		default:
			panic(fmt.Sprintf("Cannot retrieve text from frame %s.", name4))
		}
	} else { // frame not found
		return "", false
	}
}

func (me *Tag) SetTextByName4(name4 TEXT, text string) {
	if frDyn, has := me.FrameByName4(string(name4)); has {
		switch fr := frDyn.(type) {
		case *FrameText:
			if text == "" { // empty text will delete the frame
				me.DeleteFrames(func(f Frame) bool {
					return f.Name4() == string(name4)
				})
			} else {
				*fr.Text() = text
			}
		case *FrameComment:
			if text == "" { // empty text will delete the frame
				me.DeleteFrames(func(f Frame) bool {
					return f.Name4() == string(name4)
				})
			} else {
				*fr.Text() = text
			}
		default:
			panic(fmt.Sprintf("Cannot set text on frame %s.", name4))
		}

	} else { // frame does not exist yet
		var newFrame Frame // polymorphic frame
		frBase := _MakeFrameHeader(string(name4))

		if name4 == TEXT_COMMENT {
			newFrame = _NewFrameComment(frBase, "eng", "", text)
		} else {
			newFrame = _NewFrameText(frBase, text)
		}

		me.frames = append(me.frames, newFrame)
	}
}
