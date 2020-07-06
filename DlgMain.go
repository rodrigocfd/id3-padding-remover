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
	wnd        gui.WindowMain
	lstFiles   gui.ListView
	lstDetails gui.ListView
	resizer    gui.Resizer
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
		me.lstFiles.CreateReport(&me.wnd, 6, 6, 410, 318)

		me.lstDetails.CreateReport(&me.wnd, 424, 6, 232, 318)

		me.resizer.Add(&me.lstFiles, gui.RESZ_RESIZE, gui.RESZ_RESIZE).
			Add(&me.lstDetails, gui.RESZ_REPOS, gui.RESZ_RESIZE)

		return 0
	})

	me.wnd.OnMsg().WmSize(func(p wm.Size) {
		me.resizer.Adjust(p)
	})
}
