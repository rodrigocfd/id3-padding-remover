package main

import (
	"wingows/win"
)

func (me *DlgMain) eventsLstFiles() {
	me.wnd.OnMsg().LvnInsertItem(LST_FILES, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount())
	})

	me.wnd.OnMsg().LvnItemChanged(LST_FILES, func(p *win.NMLISTVIEW) {
		if !me.lstFilesSelLocked {
			me.lstFilesSelLocked = true

			go func() {
				win.Sleep(50) // wait between LVM_ITEMCHANGED updates

				me.wnd.RunUiThread(func() {
					me.updateTitlebarCount(me.lstFiles.ItemCount())
					me.displayTagsOfSelectedFiles()
					me.lstFilesSelLocked = false
				})
			}()
		}
	})

	me.wnd.OnMsg().LvnDeleteItem(LST_FILES, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount() - 1) // notification is sent before deletion

		delItem := me.lstFiles.Item(uint32(p.IItem))
		delete(me.cachedTags, delItem.Text()) // remove tag from cache
	})
}
