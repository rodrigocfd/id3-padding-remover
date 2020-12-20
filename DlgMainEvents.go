package main

import (
	"fmt"
	"runtime"
	"strings"
	"windigo/co"
	"windigo/ui"
	"windigo/win"
)

func (me *DlgMain) eventsMain() {
	me.wnd.On().WmCreate(func(_ *win.CREATESTRUCT) int {
		// MP3 files list view creation.
		me.lstFiles.
			Create(ui.Pos{X: 6, Y: 6}, ui.Size{Cx: 438, Cy: 346},
				co.LVS_REPORT|co.LVS_NOSORTHEADER|co.LVS_SHOWSELALWAYS,
				co.LVS_EX_FULLROWSELECT).
			SetContextMenu(me.lstFilesMenu).
			SetImageList(co.LVSIL_SMALL, me.iconImgList)
		me.lstFiles.Columns().AddMany([]string{"File", "Padding"}, []int{1, 60}).
			Columns().Get(0).SetWidthToFill()

		// Tag values list view creation.
		me.lstValues.
			Create(ui.Pos{X: 450, Y: 6}, ui.Size{Cx: 242, Cy: 346},
				co.LVS_REPORT|co.LVS_NOSORTHEADER,
				co.LVS_EX_GRIDLINES).
			Columns().AddMany([]string{"Field", "Value"}, []int{50, 1}).
			Columns().Get(1).SetWidthToFill()
		me.lstValues.Hwnd().EnableWindow(false)

		// Other stuff.
		me.statusBar.Create().
			Parts().AddFixed(200).
			Parts().AddFixed(200).
			Parts().AddFixed(200).
			Parts().AddResizable(1).
			Parts().
			SetTexts(
				"Alloc: 0 MB",
				"Accum alloc: 0 MB",
				"Obtained: 0 MB",
				"GC cycles: 0",
			)

		me.resizer.Add(ui.RESZ_RESIZE, ui.RESZ_RESIZE, me.lstFiles).
			Add(ui.RESZ_REPOS, ui.RESZ_RESIZE, me.lstValues)

		// Memory stats timer.
		me.wnd.Hwnd().SetTimer(TIMER_MEMSTATS, 200,
			func(msElapsed uint32) {
				m := runtime.MemStats{}
				runtime.ReadMemStats(&m)

				me.statusBar.Parts().SetTexts(
					fmt.Sprintf("Alloc: %.2f MB", float32(m.Alloc)/1024/1024),
					fmt.Sprintf("Accum alloc: %.2f MB", float32(m.TotalAlloc)/1024/1024),
					fmt.Sprintf("Obtained: %.2f MB", float32(m.Sys)/1024/1024),
					fmt.Sprintf("GC cycles: %d", m.NumGC),
				)
			})

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

		me.statusBar.ResizeToFitParent(p)
	})

	me.wnd.On().WmCommandAccelMenu(int(co.MBID_CANCEL), func(_ ui.WmCommand) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on ESC
	})

	me.wnd.On().WmDropFiles(func(p ui.WmDropFiles) {
		droppedFiles := p.RetrieveAll()
		droppedMp3s := make([]string, 0, len(droppedFiles))

		for _, path := range droppedFiles {
			if ui.Path.IsFolder(path) { // if a folder, add all MP3 directly within
				if subFiles, err := ui.Path.ListFilesInFolder(path + "\\*.mp3"); err != nil {
					panic(err.Error())
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
