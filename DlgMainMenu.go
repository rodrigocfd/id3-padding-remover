package main

import (
	"wingows/co"
	"wingows/gui"
	"wingows/gui/wm"
)

func (me *DlgMain) buildMenuAndAccel() {
	me.lstFilesMenu.CreatePopup().
		AddItem("OPEN", "&Open files...\tCtrl+O").
		AddItem("DELETE", "&Delete from list\tDel").
		AddSeparator().
		AddItem("REMPAD", "Remove &padding").
		AddItem("REMRG", "Remove Replay&Gain").
		AddItem("REMRGPIC", "Remove ReplayGain and P&ic").
		AddSeparator().
		AddItem("ABOUT", "&About...\tF1")

	accelTable := gui.AccelTable{}
	accelTable.
		Add("CTRLO", 'O', co.ACCELF_CONTROL, me.lstFilesMenu.Item("OPEN").CmdId()).
		Add("DEL", co.VK_DELETE, co.ACCELF_NONE, me.lstFilesMenu.Item("DELETE").CmdId()).
		Add("F1", co.VK_F1, co.ACCELF_NONE, me.lstFilesMenu.Item("ABOUT").CmdId())
	me.wnd.Setup().HAccel = accelTable.Build()
}

func (me *DlgMain) menuEvents() {
	me.wnd.OnMsg().WmInitMenuPopup(func(p wm.InitMenuPopup) {
		if p.Hmenu() == me.lstFilesMenu.Hmenu() {
			me.lstFilesMenu.EnableMany(me.lstFiles.SelectedItemCount() > 0,
				[]string{"DELETE", "REMPAD", "REMRG", "REMRGPIC"})
		}
	})

	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("OPEN").CmdId(), func(p wm.Command) {
		ok, mp3s := gui.ShowFileOpenMany(&me.wnd,
			[]string{"MP3 audio files (*.mp3)|*.mp3"})
		if ok {
			me.addFilesIfNotYet(mp3s)
		}
	})

	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("DELETE").CmdId(), func(p wm.Command) {
		selItems := me.lstFiles.NextItemAll(co.LVNI_SELECTED)
		me.lstFiles.SetRedraw(false)
		for i := len(selItems) - 1; i >= 0; i-- {
			selItems[i].Delete() // will fire LVM_DELETEITEM
		}
		me.lstFiles.SetRedraw(true)
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
		me.wnd.Hwnd().MessageBox(
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Wingows library.",
			"About", co.MB_ICONINFORMATION)
	})
}
