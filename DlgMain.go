package main

import (
	"fmt"
	"id3-fit/id3"
	"wingows/co"
	"wingows/ui"
	"wingows/win"
)

func main() {
	dlgMain := DlgMain{}
	dlgMain.RunAsMain()
}

type DlgMain struct {
	wnd               ui.WindowMain
	iconImgList       ui.ImageList // for system MP3 icon
	lstFiles          ui.ListView
	lstFilesMenu      ui.Menu // files list view right-click menu
	lstFilesSelLocked bool    // LVN_ITEMCHANGED is scheduled to fire
	lstValues         ui.ListView
	resizer           ui.Resizer
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

	me.eventsMain()
	me.eventsLstFiles()
	me.eventsMenu()
	return me.wnd.RunAsMain()
}

func (me *DlgMain) addFilesToListIfNotYet(mp3s []string) {
	me.lstFiles.SetRedraw(false)

	for _, mp3 := range mp3s {
		if me.lstFiles.FindItem(mp3) == nil { // not yet in the list
			tag := &id3.Tag{}
			err := tag.ReadFromFile(mp3)
			if err != nil { // error when parsing the tag
				ui.SysDlgUtil.MsgBox(&me.wnd,
					fmt.Sprintf("File:\n%s\n\n%s", mp3, err.Error()),
					"Error", co.MB_ICONERROR)
			} else {
				me.lstFiles.AddItemWithIcon(mp3, 0). // will fire LVN_INSERTITEM
									SetSubItemText(1, fmt.Sprintf("%d", tag.PaddingSize()))

				me.cachedTags[mp3] = tag // cache the tag
			}
		}
	}
	me.lstFiles.SetRedraw(true).
		Column(0).FillRoom()
}

func (me *DlgMain) displayTagsOfSelectedFiles() {
	me.lstValues.SetRedraw(false).
		DeleteAllItems() // clear all tag displays

	selFiles := me.lstFiles.SelectedItemTexts(0)

	if len(selFiles) > 1 { // multiple files selected, no tags are shown
		me.lstValues.AddItem("").
			SetSubItemText(1, fmt.Sprintf("%d selected...", len(selFiles)))

	} else if len(selFiles) == 1 { // only 1 file selected, we display its tag
		tag := me.cachedTags[selFiles[0]]

		for _, frame := range tag.Frames() { // read each frame of the tag
			valItem := me.lstValues.AddItem(frame.Name4()) // first column displays frame name

			switch myFrame := frame.(type) {
			case *id3.FrameComment:
				valItem.SetSubItemText(1,
					fmt.Sprintf("[%s] %s", myFrame.Lang(), myFrame.Text()),
				)

			case *id3.FrameText:
				valItem.SetSubItemText(1, myFrame.Text())

			case *id3.FrameMultiText:
				valItem.SetSubItemText(1, myFrame.Texts()[0]) // 1st text
				for i := 1; i < len(myFrame.Texts()); i++ {
					me.lstValues.AddItemMultiColumn([]string{"", myFrame.Texts()[i]})
				}

			case *id3.FrameBinary:
				valItem.SetSubItemText(1,
					fmt.Sprintf("%.2f KB (%.2f%%)",
						float64(len(myFrame.BinData()))/1024, // frame size in KB
						float64(len(myFrame.BinData()))*100/ // percent of whole tag size
							float64(tag.TotalTagSize())),
				)
			}
		}

	}

	me.lstValues.SetRedraw(true).
		Column(1).FillRoom()
	me.lstValues.Hwnd().EnableWindow(len(selFiles) > 0) // if no files selected, disable lstValues
}

func (me *DlgMain) reSaveTagsOfSelectedFiles(tagProcess func(tag *id3.Tag)) {
	for _, selItem := range me.lstFiles.SelectedItems() {
		selFilePath := selItem.Text()
		tag := me.cachedTags[selFilePath]

		tagProcess(tag) // tag frames can be modified before saving

		err := tag.SerializeToFile(selFilePath) // simply rewrite tag, no padding is written
		if err != nil {
			ui.SysDlgUtil.MsgBox(&me.wnd,
				fmt.Sprintf("Failed to write tag to:\n%s\n\n%s",
					selFilePath, err.Error()),
				"Writing error", co.MB_ICONERROR)
			break
		}

		tag.ReadFromFile(selFilePath)
		me.cachedTags[selFilePath] = tag // re-cache modified tag

		selItem.SetSubItemText(1, fmt.Sprintf("%d", tag.PaddingSize())) // refresh padding size
	}

	me.displayTagsOfSelectedFiles() // refresh the frames display
}

func (me *DlgMain) updateTitlebarCount(total uint) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText("ID3 Fit")
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)",
			me.lstFiles.SelectedItemCount(), total))
	}
}
