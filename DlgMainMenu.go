package main

import (
	"fmt"
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
		AddKey(co.VK_F1, co.ACCELF_NONE, MNU_ABOUT)
}

func (me *DlgMain) eventsMenu() {
	me.wnd.OnMsg().WmInitMenuPopup(func(p gui.WmInitMenuPopup) {
		if p.Hmenu() == me.lstFilesMenu.Hmenu() {
			me.lstFilesMenu.EnableManyByCmdId(
				me.lstFiles.SelectedItemCount() > 0, // 1 or more files actually selected
				[]int32{MNU_DELETE, MNU_REMPAD, MNU_REMRG, MNU_REMRGPIC})
		}
	})

	me.wnd.OnMsg().WmCommand(MNU_OPEN, func(p gui.WmCommand) {
		mp3s, ok := gui.SysDlgUtil.FileOpenMany(&me.wnd,
			[]string{"MP3 audio files (*.mp3)|*.mp3"})
		if ok {
			me.addFilesToListIfNotYet(mp3s)
		}
	})

	me.wnd.OnMsg().WmCommand(MNU_DELETE, func(p gui.WmCommand) {
		me.lstFiles.SetRedraw(false).
			DeleteItems(me.lstFiles.SelectedItems()). // will fire LVM_DELETEITEM
			SetRedraw(true)
	})

	me.wnd.OnMsg().WmCommand(MNU_REMPAD, func(p gui.WmCommand) {
		for _, selItem := range me.lstFiles.SelectedItems() {
			selFilePath := selItem.Text()
			tag := me.cachedTags[selFilePath]

			err := tag.SerializeToFile(selFilePath) // simply rewrite tag, no padding is written
			if err != nil {
				gui.SysDlgUtil.MsgBox(&me.wnd,
					fmt.Sprintf("Failed to write tag to:\n%s\n\n%s",
						selFilePath, err.Error()),
					"Writing error", co.MB_ICONERROR)
				break
			}

			tag.ReadFromFile(selFilePath)
			me.cachedTags[selFilePath] = tag // re-cache modified tag

			selItem.SetSubItemText(1, fmt.Sprintf("%d", tag.PaddingSize())) // refresh padding size
		}

		me.displayTagsOfSelectedFiles() // refresh the frames display
	})

	me.wnd.OnMsg().WmCommand(MNU_REMRG, func(p gui.WmCommand) {
		println("Remove ReplayGain")
	})

	me.wnd.OnMsg().WmCommand(MNU_REMRGPIC, func(p gui.WmCommand) {
		println("Remove ReplayGain and pic, bro")
	})

	me.wnd.OnMsg().WmCommand(MNU_ABOUT, func(p gui.WmCommand) {
		gui.SysDlgUtil.MsgBox(&me.wnd,
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Wingows library.",
			"About", co.MB_ICONINFORMATION)
	})
}
