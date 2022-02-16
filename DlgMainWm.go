package main

import (
	"fmt"
	"id3fit/prompt"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgMain) eventsWm() {
	me.wnd.On().WmInitDialog(func(_ wm.InitDialog) bool {
		// File icon image list.
		// ListView doesn't have LVS_SHAREIMAGELISTS, so it'll be automatically destroyed.
		hImgList := win.ImageListCreate(16, 16, co.ILC_COLOR32, 1, 1)
		hImgList.AddIconFromShell("mp3")
		me.lstMp3s.SetImageList(co.LVSIL_SMALL, hImgList)

		// MP3 files list view creation.
		me.lstMp3s.Columns().Add([]int{1, 60}, "File", "Padding")
		me.lstMp3s.Columns().SetWidthToFill(0)

		// Tag values list view creation.
		me.lstFrames.SetExtendedStyle(true, co.LVS_EX_GRIDLINES)
		me.lstFrames.Columns().Add([]int{50, 1}, "Field", "Value")
		me.lstFrames.Columns().SetWidthToFill(1)
		me.lstFrames.Hwnd().EnableWindow(false)

		// Status bar parts.
		me.statusBar.Parts().AddResizable(1, 1, 1, 1, 1)
		me.updateMemoryStatus()

		return true
	})

	me.wnd.On().WmSize(func(_ wm.Size) {
		me.lstMp3s.Columns().SetWidthToFill(0)
		me.lstFrames.Columns().SetWidthToFill(1)
	})

	me.wnd.On().WmCommandAccelMenu(int(co.ID_CANCEL), func(_ wm.Command) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on ESC
	})

	me.wnd.On().WmDropFiles(func(p wm.DropFiles) {
		droppedFiles := p.Hdrop().ListFilesAndFinish()
		droppedMp3s := make([]string, 0, len(droppedFiles)) // MP3 files effectively found

		for _, path := range droppedFiles {
			if win.Path.IsFolder(path) { // if a folder, add all MP3 directly within
				if subFiles, err := win.Path.ListFilesInFolder(path + "\\*.mp3"); err != nil {
					panic(err) // should really never happen
				} else {
					droppedMp3s = append(droppedMp3s, subFiles...)
				}
			} else if win.Path.HasExtension(path, ".mp3") { // not a folder, just a file
				droppedMp3s = append(droppedMp3s, path)
			}
		}

		if len(droppedMp3s) == 0 { // no MP3 files have been drag n' dropped
			prompt.Error(me.wnd, "No files added", win.StrOptVal("No files"),
				fmt.Sprintf("%d items dropped, no MP3 found.", len(droppedFiles)))
		} else {
			if tagOpErr := me.modalTagOp(droppedMp3s, TAG_OP_LOAD); tagOpErr != nil {
				prompt.Error(me.wnd, "Tag operation error", nil,
					fmt.Sprintf("Failed to load tag:\n%sn\n\n%s",
						tagOpErr.mp3, tagOpErr.err.Error()))
			} else {
				me.addMp3sToList(droppedMp3s)
			}
		}

		me.updateMemoryStatus()
	})

	me.dlgFields.OnSave(func(t0 timecount.TimeCount) {
		// Tags have been modified, but not saved to disk yet.
		// Here we save them and reload the cache.

		selMp3s := me.lstMp3s.Columns().SelectedTexts(0)

		if tagOpErr := me.modalTagOp(selMp3s, TAG_OP_SAVE_AND_RELOAD); tagOpErr != nil {
			prompt.Error(me.wnd, "Tag operation error", nil,
				fmt.Sprintf("Failed to remove padding:\n%sn\n\n%s",
					tagOpErr.mp3, tagOpErr.err.Error()))
		} else {
			me.addMp3sToList(selMp3s)
			me.displayFramesOfSelectedFiles()
			prompt.Info(me.wnd, "Process finished", win.StrOptVal("Success"),
				fmt.Sprintf("%d file(s) saved in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})
}
