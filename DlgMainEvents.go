package main

import (
	"fmt"
	"id3fit/prompt"

	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgMain) eventsMain() {
	me.wnd.On().WmCreate(func(_ wm.Create) int {
		// File icon image list.
		hImgList := win.ImageListCreate(16, 16, co.ILC_COLOR32, 1, 1)
		hImgList.AddIconFromShell("mp3")
		me.lstFiles.SetImageList(co.LVSIL_SMALL, hImgList)

		// MP3 files list view creation.
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

		return 0
	})

	me.wnd.On().WmSize(func(_ wm.Size) {
		me.lstFiles.Columns().SetWidthToFill(0)
		me.lstValues.Columns().SetWidthToFill(1)
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
			} else if win.Path.HasExtension(path, ".mp3") { // not a folder, just a file
				droppedMp3s = append(droppedMp3s, path)
			}
		}

		if len(droppedMp3s) == 0 { // no MP3 files have been drag n' dropped
			prompt.Error(me.wnd, "No files added",
				fmt.Sprintf("%d items dropped, no MP3 found.", len(droppedFiles)))
		} else {
			me.addFilesToList(droppedMp3s)
		}
	})
}
