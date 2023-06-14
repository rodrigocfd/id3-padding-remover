package dlgmain

import (
	"fmt"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui"
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
		me.lstMp3s.Columns().Get(0).SetWidthToFill()

		// Tag values list view creation.
		me.lstFrames.SetExtendedStyle(true, co.LVS_EX_GRIDLINES)
		me.lstFrames.Columns().Add([]int{50, 1}, "Field", "Value")
		me.lstFrames.Columns().Get(1).SetWidthToFill()
		me.lstFrames.Hwnd().EnableWindow(false)

		// Status bar parts.
		me.statusBar.Parts().AddResizable(1, 1, 1, 1, 1)
		me.updateMemoryStatus()

		return true
	})

	me.wnd.On().WmSize(func(_ wm.Size) {
		me.lstMp3s.Columns().Get(0).SetWidthToFill()
		me.lstFrames.Columns().Get(1).SetWidthToFill()
	})

	me.wnd.On().WmCommandAccelMenu(int(co.ID_CANCEL), func(_ wm.Command) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on ESC
	})

	me.wnd.On().WmInitMenuPopup(func(p wm.InitMenuPopup) {
		switch p.Hmenu() {
		case me.lstMp3s.ContextMenu():
			me.initMenuPopupFiles(p)
		case me.lstFrames.ContextMenu():
			me.initMenuPopupFrames(p)
		}
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
			ui.TaskDlg.Error(me.wnd, "No files added", win.StrOptSome("No files"),
				fmt.Sprintf("%d items dropped, no MP3 found.", len(droppedFiles)))
		} else {
			if me.modalTagOp(droppedMp3s, TAG_OP_LOAD) {
				me.addMp3sToList(droppedMp3s)
			}
		}

		me.updateMemoryStatus()
	})

	me.dlgFields.OnSave(func(t0 timecount.TimeCount) {
		// Tags have been modified, but not saved to disk yet.
		// Here we save them and reload the cache.

		selMp3s := me.lstMp3s.Columns().Get(0).SelectedTexts()

		if me.modalTagOp(selMp3s, TAG_OP_SAVE_AND_RELOAD) {
			me.addMp3sToList(selMp3s)
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d file(s) saved in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})
}
