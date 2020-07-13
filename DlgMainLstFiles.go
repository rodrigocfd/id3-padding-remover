package main

import (
	"wingows/win"
)

func (me *DlgMain) lstFilesEvents() {
	me.wnd.OnMsg().LvnInsertItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount())
	})

	me.wnd.OnMsg().LvnItemChanged(&me.lstFiles, func(p *win.NMLISTVIEW) {
		if !me.lstFilesSelChanging {
			me.lstFilesSelChanging = true
			go func() {
				win.Sleep(100)

				me.wnd.RunUiThread(func() {
					me.updateTitlebarCount(me.lstFiles.ItemCount())
					me.lstValues.DeleteAllItems()

					me.lstFilesSelChanging = false
				})
			}()
		}
	})

	me.wnd.OnMsg().LvnDeleteItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount() - 1) // notification is sent before deletion
	})
}
