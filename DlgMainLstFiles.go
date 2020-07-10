package main

import (
	"wingows/co"
	"wingows/win"
)

func (me *DlgMain) lstFilesEvents() {
	me.wnd.OnMsg().LvnInsertItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount())
	})

	me.wnd.OnMsg().LvnItemChanged(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount())
		me.lstValues.DeleteAllItems()

		sels := me.lstFiles.NextItemAll(co.LVNI_SELECTED)
		for _, sel := range sels {
			println(sel.Text())
		}
	})

	me.wnd.OnMsg().LvnDeleteItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount() - 1) // notification is sent before deletion
	})
}
