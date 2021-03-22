package main

import (
	"fmt"
	"id3fit/id3"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

type ParseRes struct {
	Mp3 string
	Tag *id3.Tag
	Err error
}

func (me *DlgMain) addFilesToList(mp3s []string) {
	channel := make(chan ParseRes, len(mp3s))
	for _, mp3 := range mp3s {
		theMp3 := mp3
		go func() {
			tag, lerr := id3.ParseTagFromFile(theMp3)
			channel <- ParseRes{theMp3, tag, lerr}
		}()
	}

	for i := 0; i < len(mp3s); i++ {
		parseRes := <-channel
		if parseRes.Err != nil { // if error, simply popup and move on
			ui.Native.MessageBox(me.wnd,
				fmt.Sprintf("File:\n%s\n\n%s", parseRes.Mp3, parseRes.Err),
				"Error parsing tag", co.MB_ICONERROR)
		}

		if _, found := me.lstFiles.Items().Find(parseRes.Mp3); !found { // file not yet in the list?
			me.lstFiles.Items().
				AddWithIcon(0, parseRes.Mp3,
					fmt.Sprintf("%d", parseRes.Tag.PaddingSize())) // will fire LVN_INSERTITEM
		}

		me.cachedTags[parseRes.Mp3] = parseRes.Tag // cache (or re-cache) the tag
	}

	me.lstFiles.Columns().SetWidthToFill(0)
}

func (me *DlgMain) displayTagsOfSelectedFiles() {
	me.lstValues.SetRedraw(false)
	me.lstValues.Items().DeleteAll() // clear all tag displays

	selPaths := me.lstFiles.Columns().SelectedTexts(0)

	if len(selPaths) > 1 { // multiple files selected, no tags are shown
		me.lstValues.Items().
			Add("", fmt.Sprintf("%d selected...", len(selPaths)))

	} else if len(selPaths) == 1 { // only 1 file selected, we display its tag
		cachedTag := me.cachedTags[selPaths[0]]

		for _, frameDyn := range cachedTag.Frames() { // read each frame of the tag
			newValIdx := me.lstValues.Items().
				Add(frameDyn.Name4()) // add new item, first column displays frame name

			switch myFrame := frameDyn.(type) {
			case *id3.FrameComment:
				me.lstValues.Items().SetText(newValIdx, 1,
					fmt.Sprintf("[%s] %s", myFrame.Lang(), myFrame.Text()))

			case *id3.FrameText:
				me.lstValues.Items().SetText(newValIdx, 1, myFrame.Text())

			case *id3.FrameMultiText:
				me.lstValues.Items().SetText(newValIdx, 1, myFrame.Texts()[0]) // 1st text
				for i := 1; i < len(myFrame.Texts()); i++ {
					me.lstValues.Items().Add("", myFrame.Texts()[i]) // subsequent
				}

			case *id3.FrameBinary:
				binLen := uint64(len(myFrame.BinData()))
				me.lstValues.Items().SetText(newValIdx, 1,
					fmt.Sprintf("%s (%.2f%%)",
						win.Str.FmtBytes(binLen), // frame size
						float64(binLen)*100/ // percent of whole tag size
							float64(cachedTag.TotalTagSize())),
				)
			}
		}

	}

	me.lstValues.SetRedraw(true)
	me.lstValues.Columns().SetWidthToFill(1)
	me.lstValues.Hwnd().EnableWindow(len(selPaths) > 0) // if no files selected, disable lstValues
}

func (me *DlgMain) reSaveTagsOfSelectedFiles(
	tagProcessBeforeSave func(tag *id3.Tag)) {

	for _, selIdx := range me.lstFiles.Items().Selected() {
		selFilePath := me.lstFiles.Items().Text(selIdx, 0)
		tag := me.cachedTags[selFilePath]

		tagProcessBeforeSave(tag) // tag frames can be modified before saving

		if err := tag.SerializeToFile(selFilePath); err != nil { // simply rewrite tag, no padding is written
			ui.Native.MessageBox(me.wnd,
				fmt.Sprintf("Failed to write tag to:\n%s\n\n%s",
					selFilePath, err.Error()),
				"Writing error", co.MB_ICONERROR)
			break
		}

		reTag, err := id3.ParseTagFromFile(selFilePath) // parse newly saved tag
		if err != nil {
			ui.Native.MessageBox(me.wnd,
				fmt.Sprintf("Failed to rescan saved file:\n%s\n\n%s", selFilePath, err.Error()),
				"Error", co.MB_ICONERROR)
			break
		}

		me.cachedTags[selFilePath] = reTag // re-cache modified tag

		me.lstFiles.Items().SetText(selIdx, 1,
			fmt.Sprintf("%d", tag.PaddingSize())) // refresh padding size
	}

	me.displayTagsOfSelectedFiles() // refresh the frames display
}

func (me *DlgMain) updateTitlebarCount(total int) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText("ID3 Fit")
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)",
			me.lstFiles.Items().SelectedCount(), total))
	}
}
