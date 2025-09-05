//go:build windows

package id3v2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/xslices"
)

// Each MP3 file has a single ID3v2 tag.
type Tag struct {
	mp3Offset int
	padding   int
	frames    []*Frame
}

func (me *Tag) Mp3Offset() int   { return me.mp3Offset }
func (me *Tag) Padding() int     { return me.padding }
func (me *Tag) Frames() []*Frame { return me.frames }

// Constructor.
func TagFromFile(mp3Path string) (*Tag, error) {
	fin, err := win.FileMapOpen(mp3Path, co.FOPEN_READ_EXISTING)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	me, err := TagFromBin(fin.HotSlice())
	if err != nil {
		return nil, err
	}
	return me, nil
}

// Constructor.
func TagFromBin(src []byte) (*Tag, error) {
	me := &Tag{}

	declaredSize, err := tagParseHeader(src)
	if err != nil {
		return nil, err
	} else if declaredSize == 0 {
		return me, nil // MP3 has no ID3v2 tag
	}

	me.frames, me.mp3Offset, me.padding, err = tagParseFrames(src[10:]) // skip 10-byte tag header
	if err != nil {
		return nil, err
	}

	return me, nil
}

// If an ID3v2 tag is present, returns its declared size, including the 10-byte
// header. Otherwise, returns zero.
func tagParseHeader(src []byte) (declaredSize int, err error) {
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

	return int(nDeclaredSize), nil
}

// Returns the frames, MP3 offset and padding size.
func tagParseFrames(src []byte) (frames []*Frame, mp3Offset, padding int, err error) {
	frames = make([]*Frame, 0, 10) // arbitrary
	mp3Offset = 10                 // start at 10 because src already skipped 10-byte header

	// Known magic byte sequences that identify the beginning of a MP3.
	// https://stackoverflow.com/a/7302482/6923555
	// https://en.wikipedia.org/wiki/List_of_file_signatures
	// https://github.com/sindresorhus/file-type/issues/75#issuecomment-320650344
	MP3_MAGIC := [][2]byte{{0xff, 0xfb}, {0xff, 0xfb}, {0xff, 0xf2}, {0xff, 0xfa}, {0xff, 0xf3}}

	for {
		if slices.ContainsFunc(MP3_MAGIC, func(mp3Magic [2]byte) bool {
			return bytes.Equal(src[:2], mp3Magic[:])
		}) {
			// We found the beginning of the MP3 file, no padding.
			return frames, mp3Offset, 0, nil
		}

		if src[0] == 0x00 {
			// We entered a padding region after all frames.
			for i := 1; i < len(src)-1; i++ { // skip the 1st byte, which is 0x00; don't count last, we're checking 2
				for _, mp3Magic := range MP3_MAGIC {
					if bytes.Equal(src[i:i+2], mp3Magic[:]) {
						return frames, mp3Offset + i, i, nil
					}
				}
			}
			return nil, 0, 0, errors.New("MP3 offset not found")
		}

		pFrame, err := _FrameParse(src)
		if err != nil {
			return nil, 0, 0, err
		}

		if pFrame.DeclaredSize() > len(src) { // means the size was serialized with error
			return nil, 0, 0, fmt.Errorf("declared frame size greater than available size: %d vs %d",
				pFrame.DeclaredSize(), len(src))
		}

		mp3Offset += pFrame.DeclaredSize()
		src = src[pFrame.DeclaredSize():]
		frames = append(frames, pFrame)
	}
}

// Returns a new tag with all the data and frames copied.
func (me *Tag) Clone() *Tag {
	clonedFrames := make([]*Frame, 0, len(me.frames))
	for _, pFrame := range me.frames {
		clonedFrames = append(clonedFrames, pFrame.Clone())
	}

	return &Tag{me.mp3Offset, me.padding, clonedFrames}
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
func (me *Tag) SaveToFile(mp3Path string) error {
	fout, err := win.FileOpen(mp3Path, co.FOPEN_RW_EXISTING)
	if err != nil {
		return err
	}
	defer fout.Close()

	currentContents, err := fout.ReadAllAsVec() // copy the whole file into memory
	if err != nil {
		return err
	}
	defer currentContents.Free()

	oldTag, err := TagFromBin(currentContents.HotSlice()) // so we can have the MP3 offset
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
	apicSz := 0
	if pApic := me.FrameByName4("APIC"); pApic != nil {
		pApicBody, _ := pApic.Body().(*BodyPicture)
		apicSz = len(pApicBody.Bin)
	}

	buf := win.NewVecReserved[byte](10 + 10*len(me.frames) + apicSz) // arbitrary

	buf.Append([]byte("ID3")...) // magic bytes
	buf.Append(0x03, 0x00)       // tag version
	buf.Append(0x00)             // flags

	buf.AppendN(4, 0x00) // placeholder for body size

	framesSz := 0 // won't count 10-byte tag header
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

	allSame := xslices.EveryFunc(tags[1:], func(_ int, pTag *Tag) bool { // skip first tag
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
