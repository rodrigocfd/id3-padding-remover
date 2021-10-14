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

func (me *Tag) OriginalSize() int    { return me.originalSize }
func (me *Tag) OriginalPadding() int { return me.originalPadding }
func (me *Tag) Frames() []Frame      { return me.frames }
func (me *Tag) IsEmpty() bool        { return len(me.frames) == 0 }

// Public constructor.
func NewEmptyTag() *Tag { return &Tag{} }

// Public constructor; reads the tag from an MP3 file.
func ReadTagFromFile(mp3Path string) (*Tag, error) { return (&Tag{}).readFromFile(mp3Path) }

// Public constructor; reads the tag from a binary blob.
func ReadTagFromBinary(src []byte) (*Tag, error) { return (&Tag{}).readFromBinary(src) }

func (me *Tag) readFromFile(mp3Path string) (*Tag, error) {
	fMap, err := win.OpenFileMapped(mp3Path, co.OPEN_FILE_READ_EXISTING)
	if err != nil {
		return nil, fmt.Errorf("opening file to read: %w", err)
	}
	defer fMap.Close()

	return me.readFromBinary(fMap.HotSlice())
}

func (me *Tag) readFromBinary(src []byte) (*Tag, error) {
	originalSize, err := me.parseTagHeader(src)
	if err != nil {
		return nil, fmt.Errorf("parsing tag header: %w", err)
	}
	src = src[10:originalSize] // skip 10-byte tag header; truncate to tag bounds

	frames, originalPadding, err := me.parseAllFrames(src)
	if err != nil {
		return nil, fmt.Errorf("parsing all frames: %w", err)
	}

	me.originalSize = originalSize
	me.originalPadding = originalPadding
	me.frames = frames
	return me, nil
}

func (me *Tag) parseTagHeader(src []byte) (int, error) {
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
			return nil, 0, fmt.Errorf("parsing frames: %w", err) // error when parsing the frame
		}
		if newFrame.OriginalSize() > len(src) { // means the tag was serialized with error
			return nil, 0, fmt.Errorf("frame size is greater than real size")
		}
		frames = append(frames, newFrame) // add the frame to our collection

		src = src[newFrame.OriginalSize():] // now starts at 1st byte of next frame
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

	fout, err := win.OpenFileMapped(mp3Path, co.OPEN_FILE_RW_EXISTING)
	if err != nil {
		return fmt.Errorf("opening file to serialize: %w", err)
	}
	defer fout.Close()
	foutMem := fout.HotSlice()

	currentTag, err := ReadTagFromBinary(foutMem)
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

func (me *Tag) TextByName(name4 string) (string, bool) {
	if frDyn, has := me.FrameByName(name4); has {
		switch fr := frDyn.(type) {
		case *FrameText:
			return *fr.Text(), true
		case *FrameComment:
			return *fr.Text(), true
		default:
			return "", false // other types not considered
		}
	} else { // frame not found
		return "", false
	}
}
