package main

import (
	"fmt"
	"id3-fit/id3"
	"strings"
	"wingows/co"
	"wingows/gui"
)

func (me *DlgMain) eventsMain() {
	me.wnd.OnMsg().WmCreate(func(p gui.WmCreate) int32 {
		// Dimensions of our two list views.
		cxLstValues := uint32(222)
		cyLstValues := p.CreateStruct().Cy - 52

		cxLstFiles := uint32(510)
		cyLstFiles := cyLstValues

		// MP3 files list view creation.
		me.lstFiles.CreateSortedReport(&me.wnd, LST_FILES, 6, 6, cxLstFiles, cyLstFiles).
			SetContextMenu(&me.lstFilesMenu).
			SetImageList(co.LVSIL_SMALL, me.iconImgList.Himagelist())
		me.lstFiles.AddColumns([]string{"File", "Padding"}, []uint32{1, 60}).
			Column(0).FillRoom()

		// Tag values list view creation.
		me.lstValues.CreateReport(&me.wnd, LST_VALUES, int32(cxLstFiles)+14, 6, cxLstValues, cyLstValues).
			AddColumns([]string{"Field", "Value"}, []uint32{50, 1}).
			Column(1).FillRoom()
		me.lstValues.Hwnd().EnableWindow(false)

		// Other stuff.
		me.resizer.Add(&me.lstFiles, gui.RESZ_RESIZE, gui.RESZ_RESIZE).
			Add(&me.lstValues, gui.RESZ_REPOS, gui.RESZ_RESIZE)

		me.cachedTags = make(map[string]*id3.Tag)
		return 0
	})

	me.wnd.OnMsg().WmSize(func(p gui.WmSize) {
		me.resizer.Adjust(p)
		me.lstFiles.Column(0).FillRoom()
		me.lstValues.Column(1).FillRoom()
	})

	me.wnd.OnMsg().WmCommand(int32(co.MBID_CANCEL), func(p gui.WmCommand) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on ESC
	})

	me.wnd.OnMsg().WmDropFiles(func(p gui.WmDropFiles) {
		paths := p.RetrieveAll()
		mp3s := make([]string, 0, len(paths))

		for _, path := range paths {
			if gui.PathUtil.PathIsFolder(path) { // if a folder, add all MP3 directly within
				subFiles := gui.PathUtil.ListFilesInFolder(path + "\\*.mp3")
				mp3s = append(mp3s, subFiles...)
			} else if strings.HasSuffix(strings.ToLower(path), ".mp3") { // not a folder, just a file
				mp3s = append(mp3s, path)
			}
		}

		if len(mp3s) == 0 { // no MP3 files have been drag n' dropped
			gui.SysDlgUtil.MsgBox(&me.wnd,
				fmt.Sprintf("%d items dropped, no MP3 found.", len(paths)),
				"No files added", co.MB_ICONEXCLAMATION)
		} else {
			me.addFilesToListIfNotYet(mp3s)
		}
	})
}
