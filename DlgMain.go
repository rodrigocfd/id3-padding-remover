package main

import (
	"fmt"
	"strings"
	"wingows/co"
	"wingows/gui"
	"wingows/gui/wm"
	"wingows/win"
)

func main() {
	dlgMain := DlgMain{}
	dlgMain.RunAsMain()
}

type DlgMain struct {
	wnd          gui.WindowMain
	lstFiles     gui.ListView
	lstFilesMenu gui.Menu
	lstValues    gui.ListView
	resizer      gui.Resizer
}

func (me *DlgMain) RunAsMain() {
	defer me.wnd.RunAsMain()

	me.wnd.Setup().Title = "ID3 Fit"
	me.wnd.Setup().Style |= co.WS_MINIMIZEBOX | co.WS_MAXIMIZEBOX | co.WS_SIZEBOX
	me.wnd.Setup().ExStyle |= co.WS_EX_ACCEPTFILES
	me.wnd.Setup().Width = 740
	me.wnd.Setup().Height = 346
	me.wnd.Setup().HIcon = win.GetModuleHandle("").LoadIcon(co.IDI(101))

	me.buildMenuAndAccel()
	defer me.lstFilesMenu.Destroy()

	me.lstFilesEvents()
	me.menuEvents()

	me.wnd.OnMsg().WmCreate(func(p wm.Create) int32 {
		il := gui.ImageList{}
		il.Create(16, 1)
		il.AddShellIcon("*.mp3")

		lstFilesCx, lstFilesCy := uint32(470), uint32(294)
		me.lstFiles.CreateReport(&me.wnd, 6, 6, lstFilesCx, lstFilesCy).
			SetContextMenu(me.lstFilesMenu.Hmenu()).
			SetImageList(co.LVSIL_SMALL, il.Himagelist())
		col1 := me.lstFiles.AddColumn("File", 1)
		me.lstFiles.AddColumn("Padding", 80)
		col1.FillRoom()

		me.lstValues.CreateReport(&me.wnd, int32(lstFilesCx)+14, 6, 232, lstFilesCy)
		me.lstValues.AddColumn("Field", 100)
		me.lstValues.AddColumn("Value", 1).FillRoom()
		me.lstValues.Hwnd().EnableWindow(false)

		me.resizer.Add(&me.lstFiles, gui.RESZ_RESIZE, gui.RESZ_RESIZE).
			Add(&me.lstValues, gui.RESZ_REPOS, gui.RESZ_RESIZE)

		return 0
	})

	me.wnd.OnMsg().WmSize(func(p wm.Size) {
		me.resizer.Adjust(p)
		me.lstFiles.Column(0).FillRoom()
	})

	me.wnd.OnMsg().WmCommand(int32(co.MBID_CANCEL), func(p wm.Command) {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0) // close on Esc
	})

	me.wnd.OnMsg().WmDropFiles(func(p wm.DropFiles) {
		paths := p.RetrieveAll()
		mp3s := make([]string, 0, len(paths))
		for _, path := range paths {
			if gui.PathIsFolder(path) { // if a folder, add all MP3 directly within
				subFiles := gui.ListFilesInFolder(path + "\\*.mp3")
				mp3s = append(mp3s, subFiles...)
			} else if strings.HasSuffix(strings.ToLower(path), ".mp3") { // not a folder, just a file
				mp3s = append(mp3s, path)
			}
		}

		if len(mp3s) == 0 {
			me.wnd.Hwnd().MessageBox(
				fmt.Sprintf("%d items dropped, no MP3 found.", len(paths)),
				"No files added", co.MB_ICONEXCLAMATION)
		} else {
			me.addFilesIfNotYet(mp3s)
		}
	})
}
