package main

import (
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
	wnd       gui.WindowMain
	lstFiles  gui.ListView
	lstValues gui.ListView
	resizer   gui.Resizer
}

func (me *DlgMain) RunAsMain() {
	me.wnd.Setup().Title = "ID3 Fit"
	me.wnd.Setup().Style |= co.WS_MINIMIZEBOX | co.WS_MAXIMIZEBOX | co.WS_SIZEBOX
	me.wnd.Setup().Width = 680
	me.wnd.Setup().Height = 370
	me.wnd.Setup().HIcon = win.GetModuleHandle("").LoadIcon(co.IDI(101))

	me.events()
	me.wnd.RunAsMain()
}

func (me *DlgMain) events() {
	me.wnd.OnMsg().WmCreate(func(p wm.Create) int32 {
		il := gui.ImageList{}
		il.Create(16, 1)
		il.AddShellIcon("*.mp3")

		me.lstFiles.CreateReport(&me.wnd, 6, 6, 410, 318)
		me.lstFiles.SetImageList(co.LVSIL_SMALL, il.Himagelist())
		me.lstFiles.AddColumn("File", 1)
		me.lstFiles.AddColumn("Padding", 80)
		me.lstFiles.Column(0).FillRoom()

		me.lstValues.CreateReport(&me.wnd, 424, 6, 232, 318)
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
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
	})
}
