package main

import (
	"fmt"
	"id3fit/dlgfields"
	"id3fit/id3v2"
	"id3fit/prompt"
	"runtime"
	"strconv"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

func main() {
	runtime.LockOSThread()
	NewDlgMain().Run()
}

type DlgMain struct {
	wnd              ui.WindowMain
	lstMp3s          ui.ListView
	lstMp3sSelLocked bool // LVN_ITEMCHANGED is scheduled to fire?
	dlgFields        *dlgfields.DlgFields
	lstFrames        ui.ListView
	cachedTags       map[string]*id3v2.Tag // for each file currently in the list
}

func NewDlgMain() *DlgMain {
	wnd := ui.NewWindowMainDlg(DLG_MAIN, ICO_MAIN, ACC_MAIN)

	me := &DlgMain{
		wnd:        wnd,
		lstMp3s:    ui.NewListViewDlg(wnd, LST_MP3S, ui.HORZ_RESIZE, ui.VERT_RESIZE, MNU_MAIN),
		dlgFields:  dlgfields.NewDlgFields(wnd, win.POINT{X: 332, Y: 6}, ui.HORZ_REPOS, ui.VERT_NONE),
		lstFrames:  ui.NewListViewDlg(wnd, LST_FRAMES, ui.HORZ_REPOS, ui.VERT_RESIZE, 0),
		cachedTags: make(map[string]*id3v2.Tag),
	}

	me.eventsWm()
	me.eventsLstFiles()
	me.eventsMenu()
	return me
}

func (me *DlgMain) Run() int {
	defer me.lstMp3s.ContextMenu().DestroyMenu()

	return me.wnd.RunAsMain()
}

func (me *DlgMain) addFilesToList(mp3s []string, onFinish func()) {
	go func() { // launch a separated thread
		halted := false

		for _, mp3 := range mp3s {
			tag, err := id3v2.ReadTagFromFile(mp3) // read all files sequentially
			if _, ok := err.(*id3v2.ErrorNoTagFound); ok {
				tag = id3v2.NewEmptyTag()
			} else if err != nil {
				me.wnd.RunUiThread(func() {
					prompt.Error(me.wnd, "Error parsing tag", nil,
						fmt.Sprintf("File:\n%s\n\n%s", mp3, err))
				})
				halted = true // nothing else will be done
				break
			}
			me.cachedTags[mp3] = tag // cache (or re-cache) tag
		}

		if halted {
			return
		}

		me.wnd.RunUiThread(func() {
			me.lstMp3s.SetRedraw(false)
			for _, mp3 := range mp3s {
				tag := me.cachedTags[mp3]

				padding := "N/A"
				if !tag.IsEmpty() {
					padding = strconv.Itoa(tag.OriginalPadding())
				}

				if item, found := me.lstMp3s.Items().Find(mp3); !found { // file not added yet?
					me.lstMp3s.Items().AddWithIcon(0, mp3, padding)
				} else {
					item.SetText(1, padding) // update padding
				}
			}
			me.lstMp3s.SetRedraw(true)
			me.lstMp3s.Columns().SetWidthToFill(0)
			me.displayFramesOfSelectedFiles()

			if onFinish != nil {
				onFinish()
			}
		})
	}()
}

func (me *DlgMain) displayFramesOfSelectedFiles() {
	me.lstFrames.SetRedraw(false)
	me.lstFrames.Items().DeleteAll() // clear all tag displays

	selMp3s := me.lstMp3s.Columns().SelectedTexts(0)

	if len(selMp3s) > 1 { // multiple files selected, no tags are shown
		me.lstFrames.Items().
			Add("", fmt.Sprintf("%d selected...", len(selMp3s)))

	} else if len(selMp3s) == 1 { // only 1 file selected, we display its tag
		cachedTag := me.cachedTags[selMp3s[0]]

		for _, frameDyn := range cachedTag.Frames() { // read each frame of the tag
			newItem := me.lstFrames.Items().
				Add(frameDyn.Name4()) // add new item, first column displays frame name

			switch myFrame := frameDyn.(type) {
			case *id3v2.FrameComment:
				newItem.SetText(1,
					fmt.Sprintf("[%s] %s", *myFrame.Lang(), *myFrame.Text()))

			case *id3v2.FrameText:
				newItem.SetText(1, *myFrame.Text())

			case *id3v2.FrameMultiText:
				newItem.SetText(1, (*myFrame.Texts())[0]) // 1st text
				for i := 1; i < len(*myFrame.Texts()); i++ {
					me.lstFrames.Items().Add("", (*myFrame.Texts())[i]) // subsequent
				}

			case *id3v2.FrameBinary:
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

	me.lstFrames.SetRedraw(true)
	me.lstFrames.Columns().SetWidthToFill(1)
	me.lstFrames.Hwnd().EnableWindow(len(selMp3s) > 0) // if no files selected, disable lstValues

	selTags := make([]*id3v2.Tag, 0, len(selMp3s))
	for _, selMp3 := range selMp3s {
		selTags = append(selTags, me.cachedTags[selMp3])
	}
	me.dlgFields.Feed(selTags)
}

func (me *DlgMain) reSaveTagsOfSelectedFiles(onFinish func()) {
	selMp3s := me.lstMp3s.Columns().SelectedTexts(0)

	go func() { // launch a separated thread
		halted := false

		for _, selMp3 := range selMp3s {
			tag := me.cachedTags[selMp3]
			if err := tag.SerializeToFile(selMp3); err != nil {
				prompt.Error(me.wnd, "Writing error", nil,
					fmt.Sprintf("Failed to write tag to:\n%sn\n\n%s", selMp3, err.Error()))
				halted = true // nothing else will be done
				break
			}
		}

		if halted {
			return
		}

		me.wnd.RunUiThread(func() {
			me.addFilesToList(selMp3s, onFinish)
		})
	}()
}

func (me *DlgMain) updateTitlebarCount(total int) {
	// Total is not computed here because LVN_DELETEITEM notification is sent
	// before the item is actually deleted, so the count would be wrong.
	if total == 0 {
		me.wnd.Hwnd().SetWindowText(APP_TITLE)
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%s (%d/%d)",
			APP_TITLE, me.lstMp3s.Items().SelectedCount(), total))
	}
}

func (me *DlgMain) renameSelectedFiles(withTrackPrefix bool) (renamedCount int, e error) {
	selItems := me.lstMp3s.Items().Selected()

	for _, selItem := range selItems {
		selMp3 := selItem.Text(0)
		theTag := me.cachedTags[selMp3]

		var track string
		if withTrackPrefix {
			trackStr, has := theTag.TextByName4(id3v2.TEXT_TRACK)
			if !has {
				return 0, fmt.Errorf("track frame absent")
			}
			track = trackStr
		}

		artist, has := theTag.TextByName4(id3v2.TEXT_ARTIST)
		if !has {
			return 0, fmt.Errorf("artist frame absent")
		}

		title, has := theTag.TextByName4(id3v2.TEXT_TITLE)
		if !has {
			return 0, fmt.Errorf("title frame absent")
		}

		var newPath string
		if withTrackPrefix {
			trackNo, err := strconv.Atoi(track)
			if err != nil {
				return 0, fmt.Errorf("invalid track format: %s", track)
			}
			newPath = fmt.Sprintf("%s\\%02d %s - %s.mp3",
				win.Path.GetPath(selMp3), trackNo, artist, title)
		} else {
			newPath = fmt.Sprintf("%s\\%s - %s.mp3",
				win.Path.GetPath(selMp3), artist, title)
		}

		if newPath != selMp3 {
			delete(me.cachedTags, selMp3)
			me.cachedTags[newPath] = theTag // re-insert tag under new name
			selItem.SetText(0, newPath)     // rename list view item
			renamedCount++

			if err := win.MoveFile(selMp3, newPath); err != nil {
				return 0, fmt.Errorf("failed to rename:\n%s\nto\n%s", selMp3, newPath)
			}
		}
	}

	return
}
