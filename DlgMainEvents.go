package main

import (
	"fmt"
	"strings"
	"windigo/co"
	"windigo/ui"
	"windigo/win"
)

func (me *DlgMain) eventsMain() {
	me.wnd.On().WmCreate(func(_ *win.CREATESTRUCT) int {
		// MP3 files list view creation.
		me.lstFiles.
			CreateSortedReport(ui.Pos{X: 6, Y: 6}, ui.Size{Cx: 438, Cy: 348}).
			SetContextMenu(me.lstFilesMenu).
			SetImageList(co.LVSIL_SMALL, me.iconImgList)
		me.lstFiles.Columns().AddMany([]string{"File", "Padding"}, []int{1, 60}).
			Columns().Get(0).SetWidthToFill()

		// Tag values list view creation.
		me.lstValues.
			CreateReport(ui.Pos{X: 450, Y: 6}, ui.Size{Cx: 242, Cy: 348}).
			Columns().AddMany([]string{"Field", "Value"}, []int{50, 1}).
			Columns().Get(1).SetWidthToFill()
		me.lstValues.Hwnd().EnableWindow(false)

		// Other stuff.
		me.resizer.Add(ui.RESZ_RESIZE, ui.RESZ_RESIZE, me.lstFiles).
			Add(ui.RESZ_REPOS, ui.RESZ_RESIZE, me.lstValues)

		return 0
	})

	me.wnd.On().WmSize(func(p ui.WmSize) {
		me.lstFiles.SetRedraw(false)
		me.lstValues.SetRedraw(false)

		me.resizer.AdjustToParent(p)
		me.lstFiles.Columns().Get(0).SetWidthToFill()
		me.lstValues.Columns().Get(1).SetWidthToFill()

		me.lstFiles.SetRedraw(true)
		me.lstValues.SetRedraw(true)
	})

	me.wnd.On().WmCommandAccelMenu(int(co.MBID_CANCEL), func(_ ui.WmCommand) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on ESC
	})

	me.wnd.On().WmDropFiles(func(p ui.WmDropFiles) {
		paths := p.RetrieveAll()
		mp3s := make([]string, 0, len(paths))

		for _, path := range paths {
			if ui.Path.IsFolder(path) { // if a folder, add all MP3 directly within
				if subFiles, err := ui.Path.ListFilesInFolder(path + "\\*.mp3"); err != nil {
					panic(err.Error())
				} else {
					mp3s = append(mp3s, subFiles...)
				}
			} else if strings.HasSuffix(strings.ToLower(path), ".mp3") { // not a folder, just a file
				mp3s = append(mp3s, path)
			}
		}

		if len(mp3s) == 0 { // no MP3 files have been drag n' dropped
			ui.SysDlg.MsgBox(me.wnd,
				fmt.Sprintf("%d items dropped, no MP3 found.", len(paths)),
				"No files added", co.MB_ICONEXCLAMATION)
		} else {
			me.addFilesToListIfNotYet(mp3s)
		}
	})
}
