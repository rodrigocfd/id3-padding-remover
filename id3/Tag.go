package id3

import (
	"wingows/gui"
)

type Tag struct {
	parsed _Parser
}

func (me *Tag) Version() [3]uint16  { return me.parsed.Version() }
func (me *Tag) TagSize() uint32     { return me.parsed.TagSize() }
func (me *Tag) PaddingSize() uint32 { return me.parsed.PaddingSize() }
func (me *Tag) Frames() []Frame     { return me.parsed.Frames() }

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
	return me.parsed.Parse(src)
}

func (me *Tag) findByName4(name4 string) Frame {
	for _, fr := range me.Frames() {
		if fr.Name4() == name4 {
			return fr
		}
	}
	return nil // not found
}
