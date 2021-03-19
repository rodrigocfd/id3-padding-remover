package main

import (
	"fmt"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgMain) eventsMain() {
	me.wnd.On().WmCreate(func(_ wm.Create) int {
		// MP3 files list view creation.
		// me.lstFiles.
		// SetContextMenu(me.lstFilesMenu).
		// SetImageList(co.LVSIL_SMALL, me.iconImgList)
		me.lstFiles.Columns().Add([]int{1, 60}, "File", "Padding")
		me.lstFiles.Columns().SetWidthToFill(0)

		// Tag values list view creation.
		me.lstValues.Columns().Add([]int{50, 1}, "Field", "Value")
		me.lstValues.Columns().SetWidthToFill(1)

		me.lstValues.Hwnd().EnableWindow(false)

		// Status bar.
		me.statusBar.Parts().AddResizable(4, 2, 3, 3)
		me.statusBar.Parts().SetAllTexts(
			"Alloc: 0 MB",
			"Accum alloc: 0 MB",
			"Obtained: 0 MB",
			"GC cycles: 0",
		)

		// Resizer.
		me.resizer.Add(ui.RESZ_RESIZE, ui.RESZ_RESIZE, me.lstFiles).
			Add(ui.RESZ_REPOS, ui.RESZ_RESIZE, me.lstValues)

		me.updateMemoryStatus()
		return 0
	})

	me.wnd.On().WmSize(func(p wm.Size) {
		me.lstFiles.SetRedraw(false)
		me.lstValues.SetRedraw(false)

		me.lstFiles.Columns().SetWidthToFill(0)
		me.lstValues.Columns().SetWidthToFill(1)

		me.lstFiles.SetRedraw(true)
		me.lstValues.SetRedraw(true)
	})

	me.wnd.On().WmCommandAccelMenu(int(co.ID_CANCEL), func(_ wm.Command) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on ESC
	})

	me.wnd.On().WmDropFiles(func(p wm.DropFiles) {
		droppedFiles := p.Hdrop().GetFilesAndFinish()
		droppedMp3s := make([]string, 0, len(droppedFiles))

		for _, path := range droppedFiles {
			if win.Path.IsFolder(path) { // if a folder, add all MP3 directly within
				if subFiles, err := win.Path.ListFilesInFolder(path + "\\*.mp3"); err != nil {
					panic(err) // should really never happen
				} else {
					droppedMp3s = append(droppedMp3s, subFiles...)
				}
			} else if strings.HasSuffix(strings.ToLower(path), ".mp3") { // not a folder, just a file
				droppedMp3s = append(droppedMp3s, path)
			}
		}

		if len(droppedMp3s) == 0 { // no MP3 files have been drag n' dropped
			ui.SysDlg.MsgBox(me.wnd,
				fmt.Sprintf("%d items dropped, no MP3 found.", len(droppedFiles)),
				"No files added", co.MB_ICONEXCLAMATION)
		} else {
			me.addFilesToList(droppedMp3s)
		}
	})
}
