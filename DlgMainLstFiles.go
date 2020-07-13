package main

import (
	"id3-fit/id3"
	"wingows/co"
	"wingows/win"
)

func (me *DlgMain) lstFilesEvents() {
	me.wnd.OnMsg().LvnInsertItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount())
	})

	me.wnd.OnMsg().LvnItemChanged(&me.lstFiles, func(p *win.NMLISTVIEW) {
		if !me.lstFilesSelLocked {
			me.lstFilesSelLocked = true

			go func() {
				win.Sleep(100)

				me.wnd.RunUiThread(func() {
					me.updateTitlebarCount(me.lstFiles.ItemCount())

					selItems := me.lstFiles.NextItemAll(co.LVNI_SELECTED)
					// for _, selItem := range selItems {

					tag := id3.Tag{}
					tag.ReadFile(selItems[0].Text())

					me.lstValues.SetRedraw(false).
						DeleteAllItems()
					for i := range tag.Frames() {
						frame := &tag.Frames()[i]
						valItem := me.lstValues.AddItem(frame.Name4())
						if frame.Kind() == id3.FRAME_KIND_TEXT {
							valItem.SubItem(1).SetText(frame.Texts()[0])
						}
					}
					me.lstValues.SetRedraw(true).
						Hwnd().EnableWindow(true)

					// }

					me.lstFilesSelLocked = false
				})
			}()
		}
	})

	me.wnd.OnMsg().LvnDeleteItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount() - 1) // notification is sent before deletion
	})
}
