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
// If the tag has no frames, it means the tag itself is absent in file.
type Tag struct {
	declaredSize int
	mp3Offset    int
	padding      int
	frames       []Frame
}

func (me *Tag) DeclaredSize() int { return me.declaredSize }
func (me *Tag) Mp3Offset() int    { return me.mp3Offset }
func (me *Tag) Padding() int      { return me.padding }
func (me *Tag) Frames() []Frame   { return me.frames }
func (me *Tag) IsEmpty() bool     { return len(me.frames) == 0 }

// Constructor; creates a new tag with no frames.
// If saved, will actually remove the tag from file.
func TagNewEmpty() *Tag { return &Tag{} }

// Constructor; reads the tag from an MP3 file.
func TagReadFromFile(mp3Path string) (*Tag, error) {
	fin, err := win.FileMappedOpen(mp3Path, co.FILE_OPEN_READ_EXISTING)
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer fin.Close()

	return TagReadFromBinary(fin.HotSlice())
}

// Constructor; reads the tag from a binary blob.
func TagReadFromBinary(src []byte) (*Tag, error) {
	declaredSize, mp3Offset, err := _TagParseHeader(src)
	if err != nil {
		return nil, fmt.Errorf("binary read: %w", err)
	}

	if declaredSize == 0 && mp3Offset == 0 {
		return TagNewEmpty(), nil // file has no tag
	} else if declaredSize == 0 && mp3Offset > 0 {
		return nil, fmt.Errorf("file has no tag, but MP3 has offset")
	}

	frames, padding, err := _TagParseFrames(src[10:declaredSize])
	if err != nil {
		return nil, fmt.Errorf("binary read: %w", err)
	}

	return &Tag{
		declaredSize: declaredSize,
		mp3Offset:    mp3Offset,
		padding:      padding,
		frames:       frames,
	}, nil
}

func _TagParseHeader(src []byte) (declaredSize, mp3Offset int, e error) {
	// Read MP3 offset.
	mp3Off, has := util.FindMp3Signature(src)
	if !has {
		return 0, 0, fmt.Errorf("no MP3 signature found")
	}

	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return 0, mp3Off, nil // MP3 file has no tag
	}

	// Validate tag version 2.3.0.
	if !bytes.Equal(src[3:5], []byte{3, 0}) { // the first "2" is not stored in the tag
		return 0, 0, fmt.Errorf(
			"tag version 2.%d.%d is not supported, only 2.3.0",
			src[3], src[4])
	}

	// Validate unsupported flags.
	if (src[5] & 0b1000_0000) != 0 {
		return 0, 0, fmt.Errorf("unsynchronised tag not supported")
	} else if (src[5] & 0b0100_0000) != 0 {
		return 0, 0, fmt.Errorf("tag extended header not supported")
	}

	// Read declared tag size and MP3 offset.
	declaredTagSize := int(util.SynchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]), // also count 10-byte tag header
	) + 10)

	if declaredSize > mp3Off {
		return 0, 0, fmt.Errorf(
			"declared size is greater than MP3 offset: %d vs %d", declaredSize, mp3Off)
	}

	return declaredTagSize, mp3Off, nil
}

func _TagParseFrames(src []byte) (frames []Frame, padding int, e error) {
	frames = make([]Frame, 0, 10) // arbitrary capacity

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

// Serializes the tag into a raw []byte.
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

// Saves or removes the tag in an MP3 file.
func (me *Tag) SerializeToFile(mp3Path string) error {
	newTagBlob := []byte{} // if tag is empty, this will actually remove any existing tag
	if !me.IsEmpty() {
		if theNewTag, err := me.Serialize(); err != nil {
			return fmt.Errorf("serializing tag: %w", err)
		} else {
			newTagBlob = theNewTag
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

	diff := len(newTagBlob) - currentTag.Mp3Offset() // size difference between new/old tags

	if diff > 0 { // new tag is larger, we need to make room
		if err := fout.Resize(fout.Size() + diff); err != nil {
			return fmt.Errorf("increasing file room: %w", err)
		}
	}

	// Move the MP3 data block inside the file, back or forth.
	destPos := int(currentTag.Mp3Offset()) + diff
	srcPos := currentTag.Mp3Offset()
	copy(foutMem[destPos:], foutMem[srcPos:])

	// Copy the new tag into the file, no padding.
	copy(foutMem, newTagBlob)

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
