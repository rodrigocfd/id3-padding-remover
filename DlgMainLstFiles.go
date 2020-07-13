package main

import (
	"id3-fit/id3"
	"wingows/co"
	"wingows/gui"
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

					selItems := me.lstFiles.NextItemAll(co.LVNI_SELECTED)
					// for _, selItem := range selItems {
					file := gui.File{}
					file.OpenExistingForRead(selItems[0].Text())
					contents := file.ReadAll()
					file.Close()

					tag := id3.Tag{}
					tag.Read(contents)

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

					me.lstFilesSelChanging = false
				})
			}()
		}
	})

	me.wnd.OnMsg().LvnDeleteItem(&me.lstFiles, func(p *win.NMLISTVIEW) {
		me.updateTitlebarCount(me.lstFiles.ItemCount() - 1) // notification is sent before deletion
	})
}
