package main

import (
	"id3-fit/id3"
	"wingows/co"
	"wingows/ui"
)

func (me *DlgMain) buildMenuAndAccel() {
	me.lstFilesMenu.CreatePopup().
		AppendItem(MNU_OPEN, "&Open files...\tCtrl+O").
		AppendItem(MNU_DELETE, "&Delete from list\tDel").
		AppendSeparator().
		AppendItem(MNU_REMPAD, "Remove &padding").
		AppendItem(MNU_REMRG, "Remove Replay&Gain").
		AppendItem(MNU_REMRGPIC, "Remove ReplayGain and p&ic").
		AppendSeparator().
		AppendItem(MNU_ABOUT, "&About...\tF1")

	me.wnd.Setup().AcceleratorTable.
		AddChar('o', co.ACCELF_CONTROL, MNU_OPEN).
		AddKey(co.VK_DELETE, co.ACCELF_NONE, MNU_DELETE).
		AddKey(co.VK_F1, co.ACCELF_NONE, MNU_ABOUT)
}

func (me *DlgMain) eventsMenu() {
	me.wnd.OnMsg().WmInitMenuPopup(func(p ui.WmInitMenuPopup) {
		if p.Hmenu() == me.lstFilesMenu.Hmenu() {
			me.lstFilesMenu.EnableManyByCmdId(
				me.lstFiles.SelectedItemCount() > 0, // 1 or more files actually selected
				[]int{MNU_DELETE, MNU_REMPAD, MNU_REMRG, MNU_REMRGPIC})
		}
	})

	me.wnd.OnMsg().WmCommand(MNU_OPEN, func(p ui.WmCommand) {
		mp3s, ok := ui.SysDlgUtil.FileOpenMany(&me.wnd,
			[]string{"MP3 audio files (*.mp3)|*.mp3"})
		if ok {
			me.addFilesToListIfNotYet(mp3s)
		}
	})

	me.wnd.OnMsg().WmCommand(MNU_DELETE, func(p ui.WmCommand) {
		me.lstFiles.SetRedraw(false).
			DeleteItems(me.lstFiles.SelectedItems()). // will fire LVM_DELETEITEM
			SetRedraw(true)
	})

	me.wnd.OnMsg().WmCommand(MNU_REMPAD, func(p ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {})
	})

	me.wnd.OnMsg().WmCommand(MNU_REMRG, func(p ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			tag.DeleteReplayGainFrames()
		})
	})

	me.wnd.OnMsg().WmCommand(MNU_REMRGPIC, func(p ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			tag.DeleteReplayGainFrames()
			tag.DeleteFrames([]string{"APIC"})
		})
	})

	me.wnd.OnMsg().WmCommand(MNU_ABOUT, func(p ui.WmCommand) {
		ui.SysDlgUtil.MsgBox(&me.wnd,
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Wingows library.",
			"About", co.MB_ICONINFORMATION)
	})
}
