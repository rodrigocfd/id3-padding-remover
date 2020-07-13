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
				win.Sleep(50) // wait between LVM_ITEMCHANGED updates

				me.wnd.RunUiThread(func() {
					me.updateTitlebarCount(me.lstFiles.ItemCount())
					me.lstValues.SetRedraw(false).
						DeleteAllItems()

					selItems := me.lstFiles.NextItemAll(co.LVNI_SELECTED)

					for _, selItem := range selItems {
						tag := me.cachedTags[selItem.Text()]

						for i := range tag.Frames() { // read each frame of the tag
							frame := &tag.Frames()[i]
							valItem := me.lstValues.AddItem(frame.Name4())
							if frame.Kind() == id3.FRAME_KIND_TEXT {
								valItem.SubItem(1).SetText(frame.Texts()[0])
							}
						}

						break // TODO remove after processing many
					}

					me.lstValues.SetRedraw(true).
						Hwnd().EnableWindow(len(selItems) > 0)

					me.lstFilesSelLocked = false
				})
			}()
		}
	})

	me.wnd.OnMsg().LvnDeleteItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount() - 1) // notification is sent before deletion

		delItem := me.lstFiles.Item(uint32(p.IItem))
		delete(me.cachedTags, delItem.Text()) // remove tag from cache
	})
}
