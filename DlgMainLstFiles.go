package main

import (
	"fmt"
	"wingows/win"
)

func (me *DlgMain) lstFilesEvents() {
	me.wnd.OnMsg().LvnInsertItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.wnd.Hwnd().SetWindowText(
			fmt.Sprintf("ID3 Fit (%d)", me.lstFiles.ItemCount()))
	})

	me.wnd.OnMsg().LvnDeleteItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		var caption string
		if me.lstFiles.ItemCount() == 0 {
			caption = "ID3 Fit"
		} else {
			caption = fmt.Sprintf("ID3 Fit (%d)", me.lstFiles.ItemCount()-1) // notification is sent before deletion
		}
		me.wnd.Hwnd().SetWindowText(caption)
	})
}
