package main

import (
	"fmt"
	"id3-fit/id3"
	"runtime"
	"windigo/co"
	"windigo/ui"
	"windigo/win"
)

func (me *DlgMain) updateMemStatus() {
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	me.statusBar.Parts().SetTexts(
		fmt.Sprintf("Alloc: %s", win.Str.FmtBytes(m.Alloc)),
		fmt.Sprintf("Accum alloc: %s", win.Str.FmtBytes(m.TotalAlloc)),
		fmt.Sprintf("Obtained: %s", win.Str.FmtBytes(m.Sys)),
		fmt.Sprintf("GC cycles: %d", m.NumGC),
	)
}

func (me *DlgMain) addFilesToList(mp3s []string) {
	me.lstFiles.SetRedraw(false)

	for _, mp3 := range mp3s {
		tag, err := id3.ParseTagFromFile(mp3)
		if err != nil {
			ui.SysDlg.MsgBox(me.wnd,
				fmt.Sprintf("File:\n%s\n\n%s", mp3, err.Error()),
				"Error", co.MB_ICONERROR)
			continue // then just proceed to next file
		}

		if me.lstFiles.Items().Find(mp3) == nil { // file not yet in the list
			me.lstFiles.Items().
				AddWithIcon(0, mp3, fmt.Sprintf("%d", tag.PaddingSize())) // will fire LVN_INSERTITEM
		}

		me.cachedTags[mp3] = tag // cache (or re-cache) the tag
	}
	me.lstFiles.SetRedraw(true).
		Columns().Get(0).SetWidthToFill()

	me.displayTagsOfSelectedFiles()
}

func (me *DlgMain) displayTagsOfSelectedFiles() {
	me.lstValues.SetRedraw(false).
		Items().DeleteAll() // clear all tag displays

	selPaths := me.lstFiles.Columns().Get(0).SelectedItemsTexts()

	if len(selPaths) > 1 { // multiple files selected, no tags are shown
		me.lstValues.Items().
			Add("", fmt.Sprintf("%d selected...", len(selPaths)))

	} else if len(selPaths) == 1 { // only 1 file selected, we display its tag
		cachedTag := me.cachedTags[selPaths[0]]

		for _, frameDyn := range cachedTag.Frames() { // read each frame of the tag
			newValItem := me.lstValues.Items().
				Add(frameDyn.Name4()) // add new item, first column displays frame name

			switch myFrame := frameDyn.(type) {
			case *id3.FrameComment:
				newValItem.SetSubItemText(1,
					fmt.Sprintf("[%s] %s", myFrame.Lang(), myFrame.Text()),
				)

			case *id3.FrameText:
				newValItem.SetSubItemText(1, myFrame.Text())

			case *id3.FrameMultiText:
				newValItem.SetSubItemText(1, myFrame.Texts()[0]) // 1st text
				for i := 1; i < len(myFrame.Texts()); i++ {
					me.lstValues.Items().Add("", myFrame.Texts()[i]) // subsequent
				}

			case *id3.FrameBinary:
				newValItem.SetSubItemText(1,
					fmt.Sprintf("%.2f KB (%.2f%%)",
						float64(len(myFrame.BinData()))/1024, // frame size in KB
						float64(len(myFrame.BinData()))*100/ // percent of whole tag size
							float64(cachedTag.TotalTagSize())),
				)
			}
		}

	}

	me.lstValues.SetRedraw(true).
		Columns().Get(1).SetWidthToFill()
	me.lstValues.Hwnd().EnableWindow(len(selPaths) > 0) // if no files selected, disable lstValues
}

func (me *DlgMain) reSaveTagsOfSelectedFiles(
	tagProcessBeforeSave func(tag *id3.Tag)) {

	for _, selItem := range me.lstFiles.Items().Selected() {
		selFilePath := selItem.Text()
		tag := me.cachedTags[selFilePath]

		tagProcessBeforeSave(tag) // tag frames can be modified before saving

		if err := tag.SerializeToFile(selFilePath); err != nil { // simply rewrite tag, no padding is written
			ui.SysDlg.MsgBox(me.wnd,
				fmt.Sprintf("Failed to write tag to:\n%s\n\n%s",
					selFilePath, err.Error()),
				"Writing error", co.MB_ICONERROR)
			break
		}

		reTag, err := id3.ParseTagFromFile(selFilePath) // parse newly saved tag
		if err != nil {
			ui.SysDlg.MsgBox(me.wnd,
				fmt.Sprintf("Failed to rescan saved file:\n%s\n\n%s", selFilePath, err.Error()),
				"Error", co.MB_ICONERROR)
			break
		}

		me.cachedTags[selFilePath] = reTag                              // re-cache modified tag
		selItem.SetSubItemText(1, fmt.Sprintf("%d", tag.PaddingSize())) // refresh padding size
	}

	me.displayTagsOfSelectedFiles() // refresh the frames display
}

func (me *DlgMain) updateTitlebarCount(total int) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText("ID3 Fit")
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)",
			me.lstFiles.Items().SelectedCount(), total))
	}
}
