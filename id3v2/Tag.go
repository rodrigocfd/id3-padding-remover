package id3v2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"id3fit/id3v2/util"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

// Metadata of a single MP3 file.
type Tag struct {
	declaredSize int
	mp3Offset    int
	padding      int
	frames       []*Frame
}

func (me *Tag) DeclaredSize() int { return me.declaredSize }
func (me *Tag) Mp3Offset() int    { return me.mp3Offset }
func (me *Tag) Padding() int      { return me.padding }
func (me *Tag) Frames() []*Frame  { return me.frames }
func (me *Tag) IsEmpty() bool     { return len(me.frames) == 0 }

// Constructor; creates a new tag with no frames.
// If saved, will actually remove the tag from file.
func NewTagEmpty() *Tag {
	return &Tag{}
}

// Constructor; parses the tag from an MP3 file.
func NewTagParseFromFile(mp3Path string) (*Tag, error) {
	fin, err := win.FileMappedOpen(mp3Path, co.FILE_OPEN_READ_EXISTING)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	return NewTagParseFromBinary(fin.HotSlice())
}

// Constructor; parses the tag from a binary blob.
func NewTagParseFromBinary(src []byte) (*Tag, error) {
	declaredSize, mp3Offset, err := _TagParseHeader(src)
	if err != nil {
		return nil, err
	}

	if declaredSize == 0 && mp3Offset == 0 {
		return NewTagEmpty(), nil // file has no tag
	}

	frames, padding, err := _TagParseFrames(src[10:declaredSize])
	if err != nil {
		return nil, err
	}

	return &Tag{
		declaredSize: declaredSize,
		mp3Offset:    mp3Offset,
		padding:      padding,
		frames:       frames,
	}, nil
}

func _TagParseHeader(src []byte) (declaredSize, mp3Offset int, e error) {
	// Find MP3 offset.
	mp3Offset, has := util.FindMp3Signature(src)
	if !has {
		return 0, 0, fmt.Errorf("no MP3 signature found")
	}

	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return 0, mp3Offset, nil // MP3 file has no tag
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

	// Read declared tag size.
	declaredSize = int(util.SynchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]),
	) + 10) // also count 10-byte tag header

	if declaredSize > mp3Offset {
		return 0, 0, fmt.Errorf(
			"declared size is greater than MP3 offset: %d vs %d",
			declaredSize, mp3Offset)
	}

	return declaredSize, mp3Offset, nil
}

func _TagParseFrames(src []byte) (frames []*Frame, padding int, e error) {
	frames = make([]*Frame, 0, 10) // arbitrary

	for {
		if len(src) == 0 { // end of tag, no padding found
			break
		} else if util.IsSliceZeroed(src) { // we entered a padding region after all frames
			padding = len(src)
			break
		}

		newFrame, err := _NewFrameParse(src)
		if err != nil {
			return nil, 0, err
		}
		if newFrame.OriginalSize() > len(src) { // means the size was serialized with error
			return nil, 0, fmt.Errorf(
				"frame size is greater than available size: %d vs %d",
				newFrame.OriginalSize(), len(src))
		}

		frames = append(frames, newFrame) // add the frame to our collection
		src = src[newFrame.OriginalSize():]
	}

	return frames, padding, nil
}

// Serializes the tag into a []byte.
func (t *Tag) Serialize() []byte {
	t._apicAsLastFrame()
	serializedFrames := make([]byte, 0, len(t.frames)*30) // arbitrary
	for _, frame := range t.frames {
		serializedFrames = append(serializedFrames, frame.Serialize()...)
	}

	finalBlob := make([]byte, 0, 10+len(serializedFrames))
	finalBlob = append(finalBlob, []byte("ID3")...)      // magic bytes
	finalBlob = append(finalBlob, []byte{0x03, 0x00}...) // tag version 2.3.0
	finalBlob = append(finalBlob, 0x00)                  // flags

	synchSafeDataSize := util.SynchSafeEncode(uint32(len(serializedFrames))) // won't count 10-byte header
	finalBlob = util.Append32(finalBlob, binary.BigEndian, synchSafeDataSize)

	finalBlob = append(finalBlob, serializedFrames...)
	return finalBlob
}

func (t *Tag) _apicAsLastFrame() {
	idx, _, has := t.FrameByName4("APIC")
	if !has { // no APIC frame?
		return
	}

	numFrames := len(t.frames)
	if idx == numFrames-1 { // already last frame?
		return
	}

	t.SwapFrames(idx, numFrames-1)
}

// Saves or removes a tag in an MP3 file.
func (t *Tag) SerializeToFile(mp3Path string) error {
	newTagBlob := []byte{} // if tag is empty, this will actually remove any existing tag
	if !t.IsEmpty() {
		newTagBlob = t.Serialize()
	}

	fout, err := win.FileOpen(mp3Path, co.FILE_OPEN_RW_EXISTING)
	if err != nil {
		return fmt.Errorf("opening file to serialize: %w", err)
	}
	defer fout.Close()

	currentContents, err := fout.ReadAll() // read the whole MP3 file into a []byte
	if err != nil {
		return fmt.Errorf("reading contents before serializing: %w", err)
	}

	currentTag, err := NewTagParseFromBinary(currentContents) // parse tag currently saved in the MP3 file
	if err != nil {
		return fmt.Errorf("reading current tag: %w", err)
	}

	if err := fout.Resize(0); err != nil { // truncate MP3 file
		return fmt.Errorf("truncating file before serializing: %w", err)
	}

	if len(newTagBlob) > 0 { // is the tag non-empty?
		if _, err := fout.Write(newTagBlob); err != nil { // write new tag to MP3 file
			return fmt.Errorf("writing new tag: %w", err)
		}
	}

	if _, err := fout.Write(currentContents[currentTag.Mp3Offset():]); err != nil { // write MP3 itself
		return fmt.Errorf("writing MP3 own data: %w", err)
	}

	return nil
}

// Replaces the struct slice with another one, which will have only the chosen
// frames.
func (t *Tag) DeleteFrames(fun func(i int, f *Frame) (willDelete bool)) {
	newFrames := make([]*Frame, 0, len(t.frames))
	for idx, frame := range t.frames {
		willDelete := fun(idx, frame)
		if !willDelete { // the new slice will contain the non-deleted tags
			newFrames = append(newFrames, frame)
		}
	}
	t.frames = newFrames // throw the old one away
}

// Swaps to frames within the slice.
func (t *Tag) SwapFrames(indexA, indexB int) {
	tmp := t.frames[indexA]
	t.frames[indexA] = t.frames[indexB]
	t.frames[indexB] = tmp
}

// Retrieves the index and the frame according to its name.
func (t *Tag) FrameByName4(name4 string) (idx int, f *Frame, exists bool) {
	for i, frame := range t.frames {
		if frame.Name4() == name4 {
			return i, frame, true
		}
	}
	return -1, nil, false
}

// Retrieves the text of the given frame.
func (t *Tag) TextByFrameId(frameId FRAMETXT) (string, bool) {
	if _, frame, has := t.FrameByName4(string(frameId)); has {
		switch data := frame.data.(type) {
		case *FrameDataText:
			return data.Text, true
		case *FrameDataComment:
			return data.Text, true // for comments, we return Text, not Descr field
		default:
			panic(fmt.Sprintf("Cannot retrieve text from frame %s.", frameId))
		}
	} else { // frame not found
		return "", false
	}
}

// Sets the text of the given frame, which will be created if not existing.
func (t *Tag) SetTextByFrameId(frameId FRAMETXT, text string) {
	if _, frame, has := t.FrameByName4(string(frameId)); has { // frame already exists
		switch data := frame.data.(type) {
		case *FrameDataText:
			if text == "" { // empty text will delete the frame
				t.DeleteFrames(func(_ int, f *Frame) bool {
					return f.Name4() == string(frameId)
				})
			} else {
				data.Text = text
			}
		case *FrameDataComment:
			if text == "" { // empty text will delete the frame
				t.DeleteFrames(func(_ int, f *Frame) bool {
					return f.Name4() == string(frameId)
				})
			} else {
				data.Text = text
			}
		default: // not simple text or comment: something went wrong
			panic(fmt.Sprintf("Cannot set text on frame %s.", frameId))
		}

	} else { // frame does not exist yet
		newFrame := _NewFrameEmpty(string(frameId)) // may contain any FrameData type
		if frameId == FRAMETXT_COMMENT {
			newFrame.data = &FrameDataComment{
				Lang3: "eng",
				Text:  text,
			}
		} else {
			newFrame.data = &FrameDataText{
				Text: text,
			}
		}
		t.frames = append(t.frames, newFrame)
	}
}

// Tells whether the field has the same value across all tags.
//
// If so, returns the value itself.
func TagSameValueAcrossAll(tags []*Tag, frameId FRAMETXT) (string, bool) {
	if firstTagText, ok := tags[0].TextByFrameId(frameId); ok { // try to retrieve frame text from 1st tag
		for i := 1; i < len(tags); i++ { // run on each subsequent tag
			if otherTagText, hasFrame := tags[i].TextByFrameId(frameId); hasFrame {
				if otherTagText != firstTagText {
					return "", false // frame exists in subsequent tag, but text is different from 1st tag
				}
			} else {
				return "", false // frame absent in subsequent tag
			}
		}
		return firstTagText, true
	} else {
		return "", false // frame absent in first tag
	}
}
