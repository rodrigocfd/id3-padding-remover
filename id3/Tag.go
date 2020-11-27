package id3

import (
	"encoding/binary"
	"windigo/ui"
)

type Tag struct {
	totalTagSize int
	paddingSize  int
	frames       []Frame
}

// Constructor.
func ParseTagFromFile(mp3Path string) (*Tag, error) {
	fMap, err := ui.OpenFileMapped(mp3Path, ui.FILEMAP_MODE_R)
	if err != nil {
		return nil, err
	}
	defer fMap.Close() // HotSlice() needs the file to remain open

	return ParseTagFromBinary(fMap.HotSlice())
}

// Constructor.
func ParseTagFromBinary(src []byte) (*Tag, error) {
	totalTagSize, err := parseTagHeader(src)
	if err != nil {
		return nil, err
	}

	src = src[10:totalTagSize] // skip 10-byte tag header; truncate to tag bounds
	frames, paddingSize, err := parseAllFrames(src)
	if err != nil {
		return nil, err
	}

	return &Tag{
		totalTagSize: totalTagSize,
		paddingSize:  paddingSize,
		frames:       frames,
	}, nil
}

func (me *Tag) TotalTagSize() int { return me.totalTagSize }
func (me *Tag) PaddingSize() int  { return me.paddingSize }
func (me *Tag) Frames() []Frame   { return me.frames }

func (me *Tag) SerializeToFile(mp3Path string) error {
	// Serialize all frames.
	serializedFrames := make([][]byte, len(me.frames))
	tagSize := 0
	for i := range me.frames {
		serialized := me.frames[i].Serialize()
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

func (me *Tag) DeleteFrames(userFunc func(fr Frame) bool) {
	newFramesSlice := make([]Frame, 0, len(me.frames))

	for _, fr := range me.frames {
		willDelete := userFunc(fr)
		if !willDelete { // the new slice will contain the non-deleted tags
			newFramesSlice = append(newFramesSlice, fr)
		}
	}

	me.frames = newFramesSlice
}

func (me *Tag) FrameByName(name4 string) Frame {
	for _, fr := range me.frames {
		if fr.Name4() == name4 {
			return fr
		}
	}
	return nil // no such frame
}
