package main

import (
	"wingows/co"
	"wingows/gui"
)

func (me *DlgMain) buildMenuAndAccel() {
	me.lstFilesMenu.CreatePopup().
		AppendItem(MNU_OPEN, "&Open files...\tCtrl+O").
		AppendItem(MNU_DELETE, "&Delete from list\tDel").
		AppendSeparator().
		AppendItem(MNU_REMPAD, "Remove &padding").
		AppendItem(MNU_REMRG, "Remove Replay&Gain").
		AppendItem(MNU_REMRGPIC, "Remove ReplayGain and P&ic").
		AppendSeparator().
		AppendItem(MNU_ABOUT, "&About...\tF1")

	me.wnd.Setup().AcceleratorTable.
		AddChar('o', co.ACCELF_CONTROL, MNU_OPEN).
		AddKey(co.VK_DELETE, co.ACCELF_NONE, MNU_DELETE).
		AddKey(co.VK_F1, co.ACCELF_NONE, MNU_ABOUT).
		Build()
}

func (me *DlgMain) menuEvents() {
	me.wnd.OnMsg().WmInitMenuPopup(func(p gui.WmInitMenuPopup) {
		if p.Hmenu() == me.lstFilesMenu.Hmenu() {
			me.lstFilesMenu.EnableManyByCmdId(
				me.lstFiles.SelectedItemCount() > 0,
				[]int32{MNU_DELETE, MNU_REMPAD, MNU_REMRG, MNU_REMRGPIC})
		}
	})

	me.wnd.OnMsg().WmCommand(MNU_OPEN, func(p gui.WmCommand) {
		if mp3s, ok := gui.SysDlgUtil.FileOpenMany(&me.wnd,
			[]string{"MP3 audio files (*.mp3)|*.mp3"}); ok {

			me.addFilesIfNotYet(mp3s)
		}
	})

	me.wnd.OnMsg().WmCommand(MNU_DELETE, func(p gui.WmCommand) {
		selItems := me.lstFiles.NextItemAll(co.LVNI_SELECTED)
		me.lstFiles.SetRedraw(false)
		for i := len(selItems) - 1; i >= 0; i-- {
			selItems[i].Delete() // will fire LVM_DELETEITEM
		}
		me.lstFiles.SetRedraw(true)
	})

	me.wnd.OnMsg().WmCommand(MNU_REMPAD, func(p gui.WmCommand) {
		println("Remove padding")
	})

	me.wnd.OnMsg().WmCommand(MNU_REMRG, func(p gui.WmCommand) {
		println("Remove ReplayGain")
	})

	me.wnd.OnMsg().WmCommand(MNU_REMRGPIC, func(p gui.WmCommand) {
		println("Remove ReplayGain and pic, bro")
	})

	me.wnd.OnMsg().WmCommand(MNU_ABOUT, func(p gui.WmCommand) {
		me.wnd.Hwnd().MessageBox(
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Wingows library.",
			"About", co.MB_ICONINFORMATION)
	})
}
