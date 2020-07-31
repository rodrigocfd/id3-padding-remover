package main

import (
	"fmt"
	"id3-fit/id3"
	"wingows/co"
	"wingows/gui"
	"wingows/win"
)

func main() {
	dlgMain := DlgMain{}
	dlgMain.RunAsMain()
}

type DlgMain struct {
	wnd               gui.WindowMain
	lstFiles          gui.ListView
	lstFilesMenu      gui.Menu
	lstFilesSelLocked bool // LVN_ITEMCHANGED is scheduled to fire
	lstValues         gui.ListView
	resizer           gui.Resizer
	cachedTags        map[string]*id3.Tag // for each file currently in the list
}

func (me *DlgMain) RunAsMain() int {
	me.wnd.Setup().Title = "ID3 Fit"
	me.wnd.Setup().Style |= co.WS_MINIMIZEBOX | co.WS_MAXIMIZEBOX | co.WS_SIZEBOX
	me.wnd.Setup().ExStyle |= co.WS_EX_ACCEPTFILES
	me.wnd.Setup().Width = 770
	me.wnd.Setup().Height = 384
	me.wnd.Setup().HIcon = win.GetModuleHandle("").LoadIcon(co.IDI(101))

	me.buildMenuAndAccel()
	defer me.lstFilesMenu.Destroy()

	me.mainEvents()
	me.lstFilesEvents()
	me.menuEvents()
	return me.wnd.RunAsMain()
}

func (me *DlgMain) addFilesIfNotYet(mp3s []string) {
	me.lstFiles.SetRedraw(false)
	for _, mp3 := range mp3s {
		if me.lstFiles.FindItem(mp3) == nil { // not yet in the list
			tag := &id3.Tag{}
			err := tag.ReadFile(mp3)
			if err != nil {
				gui.SysDlgUtil.MsgBox(&me.wnd,
					fmt.Sprintf("File:\n%s\n\n%s", mp3, err.Error()),
					"Error", co.MB_ICONERROR)
			} else {
				newItem := me.lstFiles.AddItemWithIcon(mp3, 0) // will fire LVN_INSERTITEM
				me.cachedTags[mp3] = tag                       // load and cache the tag
				newItem.SubItem(1).SetText(fmt.Sprintf("%d", tag.PaddingSize()))
			}
		}
	}
	me.lstFiles.SetRedraw(true)
	me.lstFiles.Column(0).FillRoom()
}

func (me *DlgMain) displayTags() {
	me.lstValues.SetRedraw(false).
		DeleteAllItems()

	selItems := me.lstFiles.NextItemAll(co.LVNI_SELECTED)

	if len(selItems) > 1 {
		// Multiple tags: none of them will be shown.
		me.lstValues.AddItem("").
			SubItem(1).SetText(fmt.Sprintf("%d selected...", len(selItems)))

	} else if len(selItems) == 1 {
		tag := me.cachedTags[selItems[0].Text()]

		for i := range tag.Frames() { // read each frame of the tag
			frame := &tag.Frames()[i]
			valItem := me.lstValues.AddItem(frame.Name4()) // add each name4 to lstValues

			if frame.Kind() == id3.FRAME_KIND_TEXT ||
				frame.Kind() == id3.FRAME_KIND_MULTI_TEXT ||
				frame.Kind() == id3.FRAME_KIND_COMMENT {
				// String or multi-string frame types.
				valItem.SubItem(1).SetText(frame.Texts()[0])

				if frame.Kind() == id3.FRAME_KIND_MULTI_TEXT ||
					frame.Kind() == id3.FRAME_KIND_COMMENT {
					// These are multi-string frame types.
					for i := 1; i < len(frame.Texts()); i++ {
						additionalItem := me.lstValues.AddItem("") // add an empty line
						additionalItem.SubItem(1).SetText(frame.Texts()[i])
					}
				}

			} else if frame.Kind() == id3.FRAME_KIND_BINARY {
				valItem.SubItem(1).SetText(
					fmt.Sprintf("%.2f KB (%.2f%%)",
						float64(len(frame.BinData()))/1024, // frame size in KB
						float64(len(frame.BinData()))*100/ // percent of whole tag size
							float64(tag.TotalSize())),
				)
			}

		}
	}

	me.lstValues.SetRedraw(true).
		Column(1).FillRoom()
	me.lstValues.Hwnd().EnableWindow(len(selItems) > 0) // if no files selected, disable lstValues
}

func (me *DlgMain) updateTitlebarCount(total uint32) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText("ID3 Fit")
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)",
			me.lstFiles.SelectedItemCount(), total))
	}
}
