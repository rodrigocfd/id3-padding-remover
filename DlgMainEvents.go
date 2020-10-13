package main

import (
	"fmt"
	"id3-fit/id3"
	"strings"
	"windigo/co"
	"windigo/ui"
	"windigo/win"
)

func (me *DlgMain) eventsMain() {
	me.wnd.OnMsg().WmCreate(func(p *win.CREATESTRUCT) int {
		// MP3 files list view creation.
		me.lstFiles.
			CreateSortedReport(&me.wnd, LST_FILES, ui.Pos{X: 6, Y: 6}, ui.Size{Cx: 418, Cy: 328}).
			SetContextMenu(&me.lstFilesMenu).
			SetImageList(co.LVSIL_SMALL, me.iconImgList.Himagelist())
		me.lstFiles.AddColumns([]string{"File", "Padding"}, []uint{1, 60}).
			Column(0).FillRoom()

		// Tag values list view creation.
		me.lstValues.
			CreateReport(&me.wnd, LST_VALUES, ui.Pos{X: 430, Y: 6}, ui.Size{Cx: 242, Cy: 328}).
			AddColumns([]string{"Field", "Value"}, []uint{50, 1}).
			Column(1).FillRoom()
		me.lstValues.Hwnd().EnableWindow(false)

		// Other stuff.
		me.resizer.Add(&me.lstFiles, ui.RESZ_RESIZE, ui.RESZ_RESIZE).
			Add(&me.lstValues, ui.RESZ_REPOS, ui.RESZ_RESIZE)

		me.cachedTags = make(map[string]*id3.Tag)
		return 0
	})

	me.wnd.OnMsg().WmSize(func(p ui.WmSize) {
		me.lstFiles.SetRedraw(false)
		me.lstValues.SetRedraw(false)

		me.resizer.Adjust(p)
		me.lstFiles.Column(0).FillRoom()
		me.lstValues.Column(1).FillRoom()

		me.lstFiles.SetRedraw(true)
		me.lstValues.SetRedraw(true)
	})

	me.wnd.OnMsg().WmCommand(int(co.MBID_CANCEL), func(_ ui.WmCommand) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on ESC
	})

	me.wnd.OnMsg().WmDropFiles(func(p ui.WmDropFiles) {
		paths := p.RetrieveAll()
		mp3s := make([]string, 0, len(paths))

		for _, path := range paths {
			if ui.Path.PathIsFolder(path) { // if a folder, add all MP3 directly within
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
			ui.SysDlg.MsgBox(&me.wnd,
				fmt.Sprintf("%d items dropped, no MP3 found.", len(paths)),
				"No files added", co.MB_ICONEXCLAMATION)
		} else {
			me.addFilesToListIfNotYet(mp3s)
		}
	})
}
