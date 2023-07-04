package dlgmain

import (
	"errors"
	"fmt"
	"id3fit/id3v2"
	"runtime"
	"strconv"

	"github.com/rodrigocfd/windigo/win"
)

func (me *DlgMain) updateMemoryStatus() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	me.statusBar.Parts().SetAllTexts(
		fmt.Sprintf("Objects mem: %s", win.Str.FmtBytes(memStats.HeapAlloc)),
		fmt.Sprintf("Reserved sys: %s", win.Str.FmtBytes(memStats.HeapSys)),
		fmt.Sprintf("Idle spans: %s", win.Str.FmtBytes(memStats.HeapIdle)),
		fmt.Sprintf("GC cycles: %d", memStats.NumGC),
		fmt.Sprintf("Next GC: %s", win.Str.FmtBytes(memStats.NextGC)),
	)
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

		var padding string
		if tag.IsEmpty() {
			padding = "N/A"
		} else {
			padding = strconv.Itoa(tag.Padding())
		}

		if item, found := me.lstMp3s.Items().Find(mp3); !found { // file not added yet?
			me.lstMp3s.Items().AddWithIcon(0, mp3, padding)
		} else {
			item.SetText(1, padding) // file already in list; update padding
		}
	}

	me.lstMp3s.SetRedraw(true)
	me.lstMp3s.Columns().Get(0).SetWidthToFill()
	me.displayFramesOfSelectedFiles()
}

func (me *DlgMain) displayFramesOfSelectedFiles() {
	me.lstFrames.SetRedraw(false)
	me.lstFrames.Items().DeleteAll() // clear all tag displays

	selMp3s := me.lstMp3s.Columns().Get(0).SelectedTexts()

	if len(selMp3s) > 1 { // multiple files selected, no frames are shown
		me.lstFrames.Items().
			Add("", fmt.Sprintf("%d selected...", len(selMp3s)))

	} else if len(selMp3s) == 1 { // only 1 file selected, we display its frames
		cachedTag := me.cachedTags[selMp3s[0]]

		// Read each frame of the tag, and display it on the list.
		// Since operations can be made directly on the list items, the order of
		// the items in the list must match the order of the frames slice.
		for _, frame := range cachedTag.Frames() {
			newItem := me.lstFrames.Items().
				Add(frame.Name4()) // first column displays frame name

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
	me.lstFrames.Columns().Get(1).SetWidthToFill()
	me.lstFrames.Hwnd().EnableWindow(len(selMp3s) > 0) // if no files selected, disable lstValues

	selTags := make([]*id3v2.Tag, 0, len(selMp3s)) // filter the tags of currently selected files
	for _, selMp3 := range selMp3s {
		selTags = append(selTags, me.cachedTags[selMp3])
	}
	me.dlgFields.Feed(selTags)
}

func (me *DlgMain) renameSelectedFiles(withTrackPrefix bool) (renamedCount int, e error) {
	for _, selItem := range me.lstMp3s.Items().SelectedItems() {
		selMp3 := selItem.Text(0)
		theTag := me.cachedTags[selMp3] // tag of the MP3 we're going to rename

		var trackNumStr string
		if withTrackPrefix {
			if trackNumFromFrame, has := theTag.TextByFrameId(id3v2.FRAMETXT_TRACK); !has {
				return 0, errors.New("track frame absent")
			} else {
				trackNumStr = trackNumFromFrame
			}
		}

		artist, has := theTag.TextByFrameId(id3v2.FRAMETXT_ARTIST)
		if !has {
			return 0, errors.New("artist frame absent")
		}

		title, has := theTag.TextByFrameId(id3v2.FRAMETXT_TITLE)
		if !has {
			return 0, errors.New("title frame absent")
		}

		var newPath string
		if withTrackPrefix {
			if trackNumInt, err := strconv.Atoi(trackNumStr); err != nil {
				return 0, fmt.Errorf("invalid track value: %s", trackNumStr)
			} else {
				newPath = fmt.Sprintf("%s\\%02d %s - %s.mp3",
					win.Path.GetPath(selMp3), trackNumInt, artist, title)
			}
		} else {
			newPath = fmt.Sprintf("%s\\%s - %s.mp3",
				win.Path.GetPath(selMp3), artist, title)
		}

		if newPath != selMp3 { // file name actually changed?
			if err := win.MoveFile(selMp3, newPath); err != nil {
				return 0, fmt.Errorf("failed to rename:\n%s\nto\n%s", selMp3, newPath)
			}

			delete(me.cachedTags, selMp3)   // remove cached tag
			me.cachedTags[newPath] = theTag // re-insert tag under new name
			selItem.SetText(0, newPath)     // rename list view item
			renamedCount++
		}
	}
	return renamedCount, nil
}
