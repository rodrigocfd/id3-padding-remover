package main

import (
	"fmt"
)

func (me *DlgMain) addFilesIfNotYet(mp3s []string) {
	me.lstFiles.SetRedraw(false)
	for _, mp3 := range mp3s {
		if me.lstFiles.FindItem(mp3) == nil { // not yet in the list
			me.lstFiles.AddItemWithIcon(mp3, 0) // will fire LVN_INSERTITEM
		}
	}
	me.lstFiles.SetRedraw(true)
	me.lstFiles.Column(0).FillRoom()
}

func (me *DlgMain) updateTitlebarCount(total uint32) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText("ID3 Fit")
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)",
			me.lstFiles.SelectedItemCount(), total))
	}
}
