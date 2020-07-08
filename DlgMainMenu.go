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
	me.wnd.OnMsg().WmCommand(me.lstFilesMenu.Item("OPEN").CmdId(), func(p wm.Command) {
		ok, mp3s := gui.ShowFileOpenMany(&me.wnd,
			[]string{"MP3 audio files (*.mp3)|*.mp3"})
		if ok {
			me.lstFiles.SetRedraw(false)
			for _, mp3 := range mp3s {
				if me.lstFiles.FindItem(mp3) == nil { // not yet in the list
					me.lstFiles.AddItemWithIcon(mp3, 0) // will fire LVN_INSERTITEM
				}
			}
			me.lstFiles.SetRedraw(true)
		}
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
