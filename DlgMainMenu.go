package main

import "wingows/gui/wm"

func (me *DlgMain) buildMenu() {
	me.lstFilesMenu.CreatePopup().
		AddItem("OPEN", "&Open files...\tCtrl+O").
		AddItem("DELETE", "&Delete from list\tDel").
		AddSeparator().
		AddItem("REMPAD", "Remove &padding").
		AddItem("REMRG", "Remove Replay&Gain").
		AddItem("REMRGPIC", "Remove ReplayGain and P&ic").
		AddSeparator().
		AddItem("ABOUT", "&About...\tF1")
}

func (me *DlgMain) menuEvents() {
	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("OPEN").CmdId(), func(p wm.Command) {
		println("Open files")
	})

	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("DELETE").CmdId(), func(p wm.Command) {
		println("Delete from list")
	})

	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("REMPAD").CmdId(), func(p wm.Command) {
		println("Remove padding")
	})

	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("REMRG").CmdId(), func(p wm.Command) {
		println("Remove ReplayGain")
	})

	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("REMRGPIC").CmdId(), func(p wm.Command) {
		println("Remove ReplayGain and pic, bro")
	})

	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("ABOUT").CmdId(), func(p wm.Command) {
		println("About us")
	})
}
