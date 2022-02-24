package main

import (
	"fmt"
	"id3fit/id3v2"
	"runtime"
	"strconv"

	"github.com/rodrigocfd/windigo/win"
)

func (me *DlgMain) updateMemoryStatus() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	parts := me.statusBar.Parts()
	parts.SetText(0, fmt.Sprintf("Objects mem: %s", win.Str.FmtBytes(memStats.HeapAlloc)))
	parts.SetText(1, fmt.Sprintf("Reserved sys: %s", win.Str.FmtBytes(memStats.HeapSys)))
	parts.SetText(2, fmt.Sprintf("Idle spans: %s", win.Str.FmtBytes(memStats.HeapIdle)))
	parts.SetText(3, fmt.Sprintf("GC cycles: %d", memStats.NumGC))
	parts.SetText(4, fmt.Sprintf("Next GC: %s", win.Str.FmtBytes(memStats.NextGC)))
}

func (me *DlgMain) updateTitlebarCount(total int) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText(APP_TITLE)
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%s (%d/%d)",
			APP_TITLE, me.lstMp3s.Items().SelectedCount(), total))
	}
}

func (me *DlgMain) addMp3sToList(mp3s []string) {
	me.lstMp3s.SetRedraw(false)

	for _, mp3 := range mp3s {
		tag := me.cachedTags[mp3]

		padding := "N/A"
		if !tag.IsEmpty() {
			padding = strconv.Itoa(tag.Padding())
		}

		if item, found := me.lstMp3s.Items().Find(mp3); !found { // file not added yet?
			me.lstMp3s.Items().AddWithIcon(0, mp3, padding)
		} else {
			item.SetText(1, padding) // update padding
		}
	}

	me.lstMp3s.SetRedraw(true)
	me.lstMp3s.Columns().SetWidthToFill(0)
}

func (me *DlgMain) displayFramesOfSelectedFiles() {
	me.lstFrames.SetRedraw(false)
	me.lstFrames.Items().DeleteAll() // clear all tag displays

	selMp3s := me.lstMp3s.Columns().SelectedTexts(0)

	if len(selMp3s) > 1 { // multiple files selected, no tags are shown
		me.lstFrames.Items().
			Add("", fmt.Sprintf("%d selected...", len(selMp3s)))

	} else if len(selMp3s) == 1 { // only 1 file selected, we display its tag
		cachedTag := me.cachedTags[selMp3s[0]]

		for _, frame := range cachedTag.Frames() { // read each frame of the tag
			newItem := me.lstFrames.Items().
				Add(frame.Name4()) // add new item, first column displays frame name

			switch data := frame.Data().(type) {
			case *id3v2.FrameDataText:
				newItem.SetText(1, data.Text)

			case *id3v2.FrameDataUserText:
				newItem.SetText(1, fmt.Sprintf("%s / %s", data.Descr, data.Text))

			case *id3v2.FrameDataBinary:
				binLen := uint64(len(data.Data))
				newItem.SetText(1,
					fmt.Sprintf("%s: (%.2f%%)",
						win.Str.FmtBytes(binLen), // frame size
						float64(binLen)*100/ // percent of whole tag size
							float64(cachedTag.Mp3Offset())),
				)

			case *id3v2.FrameDataComment:
				newItem.SetText(1,
					fmt.Sprintf("[%s] %s", data.Lang3, data.Text))

			case *id3v2.FrameDataPicture:
				binLen := uint64(len(data.Data))
				newItem.SetText(1,
					fmt.Sprintf("%s - %s (%.2f%%)",
						data.Type.String(),
						win.Str.FmtBytes(binLen), // frame size
						float64(binLen)*100/ // percent of whole tag size
							float64(cachedTag.Mp3Offset())),
				)
			}
		}

	}

	me.lstFrames.SetRedraw(true)
	me.lstFrames.Columns().SetWidthToFill(1)
	me.lstFrames.Hwnd().EnableWindow(len(selMp3s) > 0) // if no files selected, disable lstValues

	selTags := make([]*id3v2.Tag, 0, len(selMp3s)) // filter the tags of currently selected files
	for _, selMp3 := range selMp3s {
		selTags = append(selTags, me.cachedTags[selMp3])
	}
	me.dlgFields.Feed(selTags)
}

func (me *DlgMain) renameSelectedFiles(withTrackPrefix bool) (renamedCount int, e error) {
	for _, selItem := range me.lstMp3s.Items().Selected() {
		selMp3 := selItem.Text(0)
		theTag := me.cachedTags[selMp3] // tag of the MP3 we're going to rename

		var trackNoStr string
		if withTrackPrefix {
			if trackNoConverted, has := theTag.TextByFrameId(id3v2.FRAMETXT_TRACK); !has {
				return 0, fmt.Errorf("track frame absent")
			} else {
				trackNoStr = trackNoConverted
			}
		}

		artist, has := theTag.TextByFrameId(id3v2.FRAMETXT_ARTIST)
		if !has {
			return 0, fmt.Errorf("artist frame absent")
		}

		title, has := theTag.TextByFrameId(id3v2.FRAMETXT_TITLE)
		if !has {
			return 0, fmt.Errorf("title frame absent")
		}

		var newPath string
		if withTrackPrefix {
			if trackNo, err := strconv.Atoi(trackNoStr); err != nil {
				return 0, fmt.Errorf("invalid track format: %s", trackNoStr)
			} else {
				newPath = fmt.Sprintf("%s\\%02d %s - %s.mp3",
					win.Path.GetPath(selMp3), trackNo, artist, title)
			}
		} else {
			newPath = fmt.Sprintf("%s\\%s - %s.mp3",
				win.Path.GetPath(selMp3), artist, title)
		}

		if newPath != selMp3 { // file name actually changed?
			delete(me.cachedTags, selMp3)   // remove cached tag
			me.cachedTags[newPath] = theTag // re-insert tag under new name
			selItem.SetText(0, newPath)     // rename list view item
			renamedCount++

			if err := win.MoveFile(selMp3, newPath); err != nil {
				return 0, fmt.Errorf("failed to rename:\n%s\nto\n%s", selMp3, newPath)
			}
		}
	}
	return renamedCount, nil
}
