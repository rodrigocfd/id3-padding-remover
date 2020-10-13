package main

import (
	"windigo/win"
)

func (me *DlgMain) eventsLstFiles() {
	me.wnd.OnMsg().LvnInsertItem(LST_FILES, func(_ *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount())
	})

	me.wnd.OnMsg().LvnItemChanged(LST_FILES, func(_ *win.NMLISTVIEW) {
		if !me.lstFilesSelLocked {
			me.lstFilesSelLocked = true

			me.wnd.Hwnd().SetTimer(TIMER_LSTFILES, 50, // wait between LVM_ITEMCHANGED updates
				func(msElapsed uint32) {
					me.wnd.Hwnd().KillTimer(TIMER_LSTFILES)
					me.updateTitlebarCount(me.lstFiles.ItemCount())
					me.displayTagsOfSelectedFiles()
					me.lstFilesSelLocked = false
				})
		}
	})

	me.wnd.OnMsg().LvnDeleteItem(LST_FILES, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount() - 1) // notification is sent before deletion

		delItem := me.lstFiles.Item(uint(p.IItem))
		delete(me.cachedTags, delItem.Text()) // remove tag from cache
	})
}
