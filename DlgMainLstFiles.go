package main

import (
	"windigo/win"
)

func (me *DlgMain) eventsLstFiles() {
	me.lstFiles.On().LvnInsertItem(func(_ *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.Items().Count())
		me.updateMemStatus()
	})

	me.lstFiles.On().LvnItemChanged(func(_ *win.NMLISTVIEW) {
		if !me.lstFilesSelLocked {
			me.lstFilesSelLocked = true

			me.wnd.Hwnd().SetTimer(TIMER_LSTFILES, 50, // wait between LVM_ITEMCHANGED updates
				func(msElapsed uint32) {
					me.wnd.Hwnd().KillTimer(TIMER_LSTFILES)
					me.updateTitlebarCount(me.lstFiles.Items().Count())
					me.displayTagsOfSelectedFiles()
					me.lstFilesSelLocked = false

					me.updateMemStatus()
				})
		}
	})

	me.lstFiles.On().LvnDeleteItem(func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.Items().Count() - 1) // notification is sent before deletion

		delItem := me.lstFiles.Items().Get(int(p.IItem))
		delete(me.cachedTags, delItem.Text()) // remove tag from cache

		me.updateMemStatus()
	})
}
