package main

import (
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgMain) eventsLstFiles() {
	me.lstFiles.On().LvnInsertItem(func(_ *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.Items().Count())
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
				})
		}
	})

	me.lstFiles.On().LvnDeleteItem(func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.Items().Count() - 1) // notification is sent before deletion

		delPath := me.lstFiles.Items().Text(int(p.IItem), 0)
		delete(me.cachedTags, delPath) // remove tag from cache
	})

	me.lstFiles.On().LvnKeyDown(func(p *win.NMLVKEYDOWN) {
		if p.WVKey == co.VK_DELETE {
			me.wnd.Hwnd().SendMessage(co.WM_COMMAND, win.MAKEWPARAM(uint16(MNU_DELETE), 1), 0)
		}
	})
}
