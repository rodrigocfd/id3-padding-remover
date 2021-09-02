package main

import (
	"fmt"
	"id3fit/id3"
	"id3fit/prompt"

	"github.com/rodrigocfd/windigo/win"
)

func (me *DlgMain) addFilesToList(mp3s []string) {
	go func() { // launch a goroutine right away
		for _, mp3 := range mp3s {
			tag, err := id3.ReadTagFromFile(mp3)
			if err != nil {
				me.wnd.RunUiThread(func() { // simply inform error and proceed to next mp3
					prompt.Error(me.wnd, "Error parsing tag", "",
						fmt.Sprintf("File:\n%s\n\n%s", mp3, err))
				})
				continue
			}

			me.wnd.RunUiThread(func() {
				if _, found := me.lstFiles.Items().Find(mp3); !found { // file not added yet?
					me.lstFiles.Items().
						AddWithIcon(0, mp3, fmt.Sprintf("%d", tag.OriginalPadding())) // will fire LVN_INSERTITEM
				}

				me.cachedTags[mp3] = tag // cache (or re-cache) the tag
			})
		}

		me.wnd.RunUiThread(func() {
			me.lstFiles.Columns().SetWidthToFill(0)
		})
	}()
}

func (me *DlgMain) displayTagsOfSelectedFiles() {
	me.lstValues.SetRedraw(false)
	me.lstValues.Items().DeleteAll() // clear all tag displays

	selItems := me.lstFiles.Items().Selected()

	if len(selItems) > 1 { // multiple files selected, no tags are shown
		me.lstValues.Items().
			Add("", fmt.Sprintf("%d selected...", len(selItems)))

	} else if len(selItems) == 1 { // only 1 file selected, we display its tag
		cachedTag := me.cachedTags[selItems[0].Text(0)]

		for _, frameDyn := range cachedTag.Frames() { // read each frame of the tag
			newItem := me.lstValues.Items().
				Add(frameDyn.Name4()) // add new item, first column displays frame name

			switch myFrame := frameDyn.(type) {
			case *id3.FrameComment:
				newItem.SetText(1,
					fmt.Sprintf("[%s] %s", *myFrame.Lang(), *myFrame.Text()))

			case *id3.FrameText:
				newItem.SetText(1, *myFrame.Text())

			case *id3.FrameMultiText:
				newItem.SetText(1, (*myFrame.Texts())[0]) // 1st text
				for i := 1; i < len(*myFrame.Texts()); i++ {
					me.lstValues.Items().Add("", (*myFrame.Texts())[i]) // subsequent
				}

			case *id3.FrameBinary:
				binLen := uint64(len(*myFrame.BinData()))
				newItem.SetText(1,
					fmt.Sprintf("%s (%.2f%%)",
						win.Str.FmtBytes(binLen), // frame size
						float64(binLen)*100/ // percent of whole tag size
							float64(cachedTag.OriginalSize())),
				)
			}
		}

	}

	me.lstValues.SetRedraw(true)
	me.lstValues.Columns().SetWidthToFill(1)
	me.lstValues.Hwnd().EnableWindow(len(selItems) > 0) // if no files selected, disable lstValues
}

func (me *DlgMain) reSaveTagsOfSelectedFiles() {
	for _, selItem := range me.lstFiles.Items().Selected() {
		selFilePath := selItem.Text(0)
		tag := me.cachedTags[selFilePath]

		if err := tag.SerializeToFile(selFilePath); err != nil { // simply rewrite tag, no padding is written
			prompt.Error(me.wnd, "Writing error", "",
				fmt.Sprintf("Failed to write tag to:\n%s\n\n%s", selFilePath, err.Error()))
			break
		}

		reTag, err := id3.ReadTagFromFile(selFilePath) // re-parse newly saved tag
		if err != nil {
			prompt.Error(me.wnd, "Re-parsing error", "",
				fmt.Sprintf("Failed to rescan saved file:\n%s\n\n%s", selFilePath, err.Error()))
			break
		}

		me.cachedTags[selFilePath] = reTag // re-cache modified tag
		selItem.SetText(1,
			fmt.Sprintf("%d", reTag.OriginalPadding())) // refresh padding size
	}

	me.displayTagsOfSelectedFiles() // refresh the frames display
}

func (me *DlgMain) updateTitlebarCount(total int) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText(APP_TITLE)
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%s (%d/%d)",
			APP_TITLE, me.lstFiles.Items().SelectedCount(), total))
	}
}

func (me *DlgMain) measureFileProcess(fun func()) {
	freq := float64(win.QueryPerformanceFrequency())
	t0 := float64(win.QueryPerformanceCounter())

	fun()

	prompt.Info(me.wnd, "Process finished", "Success",
		fmt.Sprintf("%d file(s) saved in %.2f ms.",
			me.lstFiles.Items().SelectedCount(),
			((float64(win.QueryPerformanceCounter())-t0)/freq)*1000,
		),
	)
}
