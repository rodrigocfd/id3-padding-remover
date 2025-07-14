//go:build windows

package id3v2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"id3fit/slices2"
	"slices"
	"strings"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

// Each MP3 file has a single ID3v2 tag.
type Tag struct {
	path      string
	mp3Offset uint
	padding   uint
	frames    []*Frame
}

func (me *Tag) Path() string     { return me.path }
func (me *Tag) Mp3Offset() uint  { return me.mp3Offset }
func (me *Tag) Padding() uint    { return me.padding }
func (me *Tag) Frames() []*Frame { return me.frames }

// Constructor.
func LoadTagFromFile(mp3Path string) (*Tag, error) {
	fin, err := win.FileMapOpen(mp3Path, co.FOPEN_READ_EXISTING)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	me, err := LoadTagFromBin(fin.HotSlice())
	if err != nil {
		return nil, err
	}
	me.path = mp3Path
	return me, nil
}

// Constructor.
func LoadTagFromBin(src []byte) (*Tag, error) {
	me := &Tag{} // note: path not set here

	declaredSize, err := me.tagParseHeader(src)
	if err != nil {
		return nil, err
	} else if declaredSize == 0 {
		return me, nil // MP3 has no ID3v2 tag
	}

	me.mp3Offset, err = me.parseFrames(src[10:]) // skip 10-byte tag header
	if err != nil {
		return nil, err
	}

	return me, nil
}

func (me *Tag) tagParseHeader(src []byte) (declaredSize uint, err error) {
	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return 0, nil // MP3 file has no tag
	}

	// Validate tag version 2.3.0.
	if !bytes.Equal(src[3:5], []byte{3, 0}) { // the first "2" is not stored in the tag
		return 0, fmt.Errorf("tag version 2.%d.%d is not supported, only 2.3.0", src[3], src[4])
	}

	// Validate unsupported flags.
	if (src[5] & 0b1000_0000) != 0 {
		return 0, errors.New("unsynchronised tag not supported")
	} else if (src[5] & 0b0100_0000) != 0 {
		return 0, errors.New("tag extended header not supported")
	}

	// Read declared tag size.
	nDeclaredSize := synchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]),
	) + 10 // also count 10-byte tag header

	return uint(nDeclaredSize), nil
}

func (me *Tag) parseFrames(src []byte) (mp3Offset uint, err error) {
	// How many bytes we forwarded since the beginning of file.
	// Starts at 10 because we receive src already skipped 10-byte tag header.
	fwdBytes := uint(10)

	MP3_START_BYTES := [...]byte{0xff, 0xfb} // https://stackoverflow.com/a/7302482/6923555

	for {
		if bytes.Equal(src[0:2], MP3_START_BYTES[:]) {
			// We found the beginning of the MP3 data.
			return fwdBytes, nil
		} else if slices2.AllEqual(src[:4], 0x00) {
			// The first 4 bytes should contain the 4-char frame name.
			// If they're all zero, it means we entered a padding region after all frames.
			idxMp3Offset := bytes.Index(src, MP3_START_BYTES[:])
			if idxMp3Offset == -1 {
				return 0, errors.New("MP3 offset not found")
			}
			me.padding = uint(idxMp3Offset)
			return fwdBytes + uint(idxMp3Offset), nil
		}

		pFrame, err := _FrameParse(src)
		if err != nil {
			return 0, err
		}

		if pFrame.DeclaredSize() > uint(len(src)) { // means the size was serialized with error
			return 0, fmt.Errorf("declared frame size greater than available size: %d vs %d",
				pFrame.DeclaredSize(), len(src))
		}

		fwdBytes += pFrame.DeclaredSize()
		src = src[pFrame.DeclaredSize():]
		me.frames = append(me.frames, pFrame)
	}
}

// Returns a new tag with all the data and frames copied.
func (me *Tag) Clone() *Tag {
	clonedFrames := make([]*Frame, 0, len(me.frames))
	for _, pFrame := range me.frames {
		clonedFrames = append(clonedFrames, pFrame.Clone())
	}

	return &Tag{me.path, me.mp3Offset, me.padding, clonedFrames}
}

// Appends a new frame with a simple text as its contents.
func (me *Tag) AddFrameWithText(name4, text string) {
	me.frames = append(me.frames, _FrameNewWithText(name4, text))
}

// Returns the tag with the given name, or nil of none.
func (me *Tag) FrameByName4(name4 string) *Frame {
	for _, pFrame := range me.frames {
		if pFrame.Name4() == name4 {
			return pFrame
		}
	}
	return nil
}

// Returns false if the tag has no frames.
func (me *Tag) IsEmpty() bool {
	return len(me.frames) == 0
}

// Removes the frames to which the predicate returns true.
func (me *Tag) RemoveFrameIf(fun func(pFrame *Frame) bool) {
	me.frames = slices.DeleteFunc(me.frames, func(pFrame *Frame) bool {
		return fun(pFrame)
	})
}

// Removes the frame at the given index.
func (me *Tag) RemoveFrame(index int) {
	me.frames = slices.Delete(me.frames, index, index+1)
}

// Returns a string resume of the ReplayGain tags, or an empty string if none.
func (me *Tag) ReplayGainStatus() string {
	hasTrack, hasAlbum := false, false

	for _, pFrame := range me.frames {
		if hasTrack && hasAlbum {
			break
		}

		if pFrame.Name4() == "TXXX" {
			if pBody, ok := pFrame.Body().(*BodyUserText); ok {
				descr := strings.ToLower(pBody.Descr)
				if strings.HasPrefix(descr, "replaygain_track_") {
					hasTrack = true
				} else if strings.HasPrefix(descr, "replaygain_album_") {
					hasAlbum = true
				}
			}
		}
	}

	if hasTrack && hasAlbum {
		return "TA"
	} else if hasTrack {
		return "T"
	} else if hasAlbum {
		return "A"
	} else {
		return ""
	}
}

// Saves the tag to the file whose path is saved in the tag object.
func (me *Tag) SaveToFile() error {
	if me.path == "" {
		return errors.New("Tag has no path")
	}

	fout, err := win.FileOpen(me.path, co.FOPEN_RW_EXISTING)
	if err != nil {
		return err
	}
	defer fout.Close()

	currentContents, err := fout.ReadAllAsVec() // copy the whole file into memory
	if err != nil {
		return err
	}
	defer currentContents.Free()

	oldTag, err := LoadTagFromBin(currentContents.HotSlice()) // so we can have the MP3 offset
	if err != nil {
		return err
	}

	if err := fout.Resize(0); err != nil { // truncate file
		return err
	}

	if len(me.frames) > 0 {
		tagBlob := me.Serialize()
		defer tagBlob.Free()
		if _, err := fout.Write(tagBlob.HotSlice()); err != nil {
			return err
		}
	}

	fout.Write(currentContents.HotSlice()[oldTag.Mp3Offset():]) // MP3 data
	me.padding = 0
	return nil
}

// Serializes the tag into raw bytes.
func (me *Tag) Serialize() win.Vec[byte] {
	apicSz := uint(0)
	if pApic := me.FrameByName4("APIC"); pApic != nil {
		pApicBody, _ := pApic.Body().(*BodyPicture)
		apicSz = uint(len(pApicBody.Bin))
	}

	buf := win.NewVecReserved[byte](10 + 10*uint(len(me.frames)) + apicSz) // arbitrary

	buf.Append([]byte("ID3")...) // magic bytes
	buf.Append(0x03, 0x00)       // tag version
	buf.Append(0x00)             // flags

	buf.AppendN(4, 0x00) // placeholder for body size

	framesSz := uint(0) // won't count 10-byte tag header
	for _, pFrame := range me.frames {
		framesSz += pFrame.Serialize(&buf)
	}

	binary.BigEndian.PutUint32(buf.HotSlice()[6:], synchSafeEncode(uint32(framesSz)))
	return buf
}

// If the frame is the same across all tags, returns it; otherwise returns nil.
func SameFrameAcrossAllTags(name4 string, tags []*Tag) *Frame {
	if len(tags) == 0 {
		return nil // no tags, no frame
	} else if len(tags) == 1 {
		return tags[0].FrameByName4(name4) // only 1 tag, direct query
	}

	pFrame0 := tags[0].FrameByName4(name4) // query the first tag right away
	if pFrame0 == nil {
		return nil // first tag doesn't have the frame, stop right now
	}

	allSame := slices2.AllTrueFunc(tags[1:], func(pTag *Tag) bool { // skip first tag
		pFrame := pTag.FrameByName4(name4)
		if pFrame == nil {
			return false // frame doesn't exist in this posterior tag
		} else {
			// Compare the textual rendering of this frame with the first frame.
			return pFrame.Body().AsText() == pFrame0.Body().AsText()
		}
	})

	if allSame {
		return pFrame0
	}
	return nil
}
