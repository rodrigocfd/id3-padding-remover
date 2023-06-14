package dlgmain

import (
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgMain) eventsLstFiles() {

	me.lstMp3s.On().LvnInsertItem(func(_ *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstMp3s.Items().Count())
	})

	me.lstMp3s.On().LvnItemChanged(func(_ *win.NMLISTVIEW) {
		if !me.lstMp3sSelLocked {
			me.lstMp3sSelLocked = true

			me.wnd.Hwnd().SetTimerCallback(50, // batch LVM_ITEMCHANGED updates
				func(timerId uintptr) {
					me.wnd.Hwnd().KillTimer(timerId)
					me.updateTitlebarCount(me.lstMp3s.Items().Count())
					me.displayFramesOfSelectedFiles()
					me.lstMp3sSelLocked = false
					me.updateMemoryStatus()
				})
		}
	})

	me.lstMp3s.On().LvnDeleteItem(func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstMp3s.Items().Count() - 1) // notification is sent before deletion

		delPath := me.lstMp3s.Items().Get(int(p.IItem)).Text(0)
		delete(me.cachedTags, delPath) // remove tag from cache
	})

	me.lstMp3s.On().LvnKeyDown(func(p *win.NMLVKEYDOWN) {
		if p.WVKey == co.VK_DELETE {
			me.wnd.Hwnd().SendMessage(co.WM_COMMAND,
				win.MAKEWPARAM(uint16(MNU_MP3_DELETE), 1), 0) // simulate menu command
		}
	})
}
