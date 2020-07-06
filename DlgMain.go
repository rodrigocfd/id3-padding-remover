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
	wnd gui.WindowMain
}

func (me *DlgMain) RunAsMain() {
	me.wnd.Setup().Title = "ID3 Fit"
	me.wnd.Setup().Style |= co.WS_MINIMIZEBOX | co.WS_MAXIMIZEBOX | co.WS_SIZEBOX
	me.wnd.Setup().HIcon = win.GetModuleHandle("").LoadIcon(co.IDI(101))

	me.events()
	me.wnd.RunAsMain()
}

func (me *DlgMain) events() {
	me.wnd.OnMsg().WmCreate(func(p wm.Create) int32 {
		return 0
	})
}
