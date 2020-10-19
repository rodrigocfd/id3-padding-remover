package main

import (
	"id3-fit/id3"
	"windigo/co"
	"windigo/ui"
)

func (me *DlgMain) buildMenuAndAccel() {
	me.lstFilesMenu.CreatePopup().
		AppendItem(MNU_OPEN, "&Open files...\tCtrl+O").
		AppendItem(MNU_DELETE, "&Delete from list\tDel").
		AppendSeparator().
		AppendItem(MNU_REM_PAD, "Remove &padding").
		AppendItem(MNU_REM_RG, "Remove Replay&Gain").
		AppendItem(MNU_REM_RG_PIC, "Remove ReplayGain and p&ic").
		AppendItem(MNU_PREFIX_YEAR, "Prefix album with &year").
		AppendSeparator().
		AppendItem(MNU_ABOUT, "&About...\tF1")

	me.wnd.Setup().AcceleratorTable.
		AddChar('o', co.ACCELF_CONTROL, MNU_OPEN).
		AddKey(co.VK_DELETE, co.ACCELF_NONE, MNU_DELETE).
		AddKey(co.VK_F1, co.ACCELF_NONE, MNU_ABOUT)
}

func (me *DlgMain) eventsMenu() {
	me.wnd.On().WmInitMenuPopup(func(p ui.WmInitMenuPopup) {
		if p.Hmenu() == me.lstFilesMenu.Hmenu() {
			me.lstFilesMenu.EnableItemsByCmdId(
				me.lstFiles.SelectedItemCount() > 0, // 1 or more files currently selected
				[]int{MNU_DELETE, MNU_PREFIX_YEAR, MNU_REM_PAD, MNU_REM_RG, MNU_REM_RG_PIC})
		}
	})

	me.wnd.On().WmCommand(MNU_OPEN, func(_ ui.WmCommand) {
		mp3s, ok := ui.SysDlg.FileOpenMany(&me.wnd,
			[]string{"MP3 audio files (*.mp3)|*.mp3"})
		if ok {
			me.addFilesToListIfNotYet(mp3s)
		}
	})

	me.wnd.On().WmCommand(MNU_DELETE, func(_ ui.WmCommand) {
		me.lstFiles.SetRedraw(false).
			DeleteItems(me.lstFiles.SelectedItems()). // will fire LVM_DELETEITEM
			SetRedraw(true)
	})

	me.wnd.On().WmCommand(MNU_REM_PAD, func(_ ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {}) // simply saving will remove the padding
	})

	me.wnd.On().WmCommand(MNU_REM_RG, func(_ ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			tag.DeleteReplayGainFrames()
		})
	})

	me.wnd.On().WmCommand(MNU_REM_RG_PIC, func(_ ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			tag.DeleteReplayGainFrames()
			tag.DeleteFrames([]string{"APIC"})
		})
	})

	me.wnd.On().WmCommand(MNU_PREFIX_YEAR, func(_ ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			tag.PrefixAlbumNameWithYear()
		})
	})

	me.wnd.On().WmCommand(MNU_ABOUT, func(_ ui.WmCommand) {
		ui.SysDlg.MsgBox(&me.wnd,
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Windigo library.",
			"About", co.MB_ICONINFORMATION)
	})
}
