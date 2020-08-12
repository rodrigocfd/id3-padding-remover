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
	iconImgList       gui.ImageList
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

	me.iconImgList.Create(16, 1).
		AddShellIcon("*.mp3")
	defer me.iconImgList.Destroy()

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
			err := tag.ReadFromFile(mp3)

			if err != nil { // error when parsing the tag
				gui.SysDlgUtil.MsgBox(&me.wnd,
					fmt.Sprintf("File:\n%s\n\n%s", mp3, err.Error()),
					"Error", co.MB_ICONERROR)
			} else {
				newItem := me.lstFiles.AddItemWithIcon(mp3, 0) // will fire LVN_INSERTITEM
				newItem.SubItem(1).SetText(fmt.Sprintf("%d", tag.PaddingSize()))

				me.cachedTags[mp3] = tag // cache the tag
			}
		}
	}
	me.lstFiles.SetRedraw(true).
		Column(0).FillRoom()
}

func (me *DlgMain) displayTags() {
	me.lstValues.SetRedraw(false).
		DeleteAllItems()

	selItems := me.lstFiles.SelectedItems()

	if len(selItems) > 1 {
		// Multiple tags: none of them will be shown.
		me.lstValues.AddItem("").
			SubItem(1).SetText(fmt.Sprintf("%d selected...", len(selItems)))

	} else if len(selItems) == 1 {
		tag := me.cachedTags[selItems[0].Text()]

		for _, frame := range tag.Frames() { // read each frame of the tag
			valItem := me.lstValues.AddItem(frame.Name4())

			if frComm, ok := frame.(*id3.FrameComment); ok { // comment frame
				valItem.SubItem(1).SetText(
					fmt.Sprintf("[%s] %s", frComm.Lang(), frComm.Text()),
				)
			} else {
				if frTxt, ok := frame.(*id3.FrameText); ok { // text frame
					valItem.SubItem(1).SetText(frTxt.Text())
				} else if frMulti, ok := frame.(*id3.FrameMultiText); ok { // multi text frame
					valItem.SubItem(1).SetText(frMulti.Texts()[0])

					for i := 1; i < len(frMulti.Texts()); i++ {
						additionalItem := me.lstValues.AddItem("") // add an empty 1st column
						additionalItem.SubItem(1).SetText(frMulti.Texts()[i])
					}
				} else if frBin, ok := frame.(*id3.FrameBinary); ok { // binary frame
					valItem.SubItem(1).SetText(
						fmt.Sprintf("%.2f KB (%.2f%%)",
							float64(len(frBin.Data()))/1024, // frame size in KB
							float64(len(frBin.Data()))*100/ // percent of whole tag size
								float64(tag.TagSize())),
					)
				}
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
