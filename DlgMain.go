package main

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/ids"
	"id3fit/prompt"
	"runtime"
	"strconv"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func main() {
	runtime.LockOSThread()
	NewDlgMain().Run()
}

type DlgMain struct {
	wnd               ui.WindowMain
	lstFiles          ui.ListView
	lstFilesSelLocked bool // LVN_ITEMCHANGED is scheduled to fire
	lstValues         ui.ListView
	cachedTags        map[string]*id3v2.Tag // for each file currently in the list
}

func NewDlgMain() *DlgMain {
	hAccel, hCtxMenu := createAccelTableAndMenu()

	wnd := ui.NewWindowMain(
		ui.WindowMainOpts().
			Title(ids.APP_TITLE).
			ClientArea(win.SIZE{Cx: 750, Cy: 320}).
			IconId(ids.ICO_MAIN).
			AccelTable(hAccel).
			WndStyles(co.WS_CAPTION | co.WS_SYSMENU | co.WS_CLIPCHILDREN |
				co.WS_BORDER | co.WS_VISIBLE | co.WS_MINIMIZEBOX |
				co.WS_MAXIMIZEBOX | co.WS_SIZEBOX).
			WndExStyles(co.WS_EX_ACCEPTFILES),
	)

	me := &DlgMain{
		wnd: wnd,
		lstFiles: ui.NewListView(wnd,
			ui.ListViewOpts().
				Position(win.POINT{X: 6, Y: 6}).
				Size(win.SIZE{Cx: 488, Cy: 306}).
				Horz(ui.HORZ_RESIZE).
				Vert(ui.VERT_RESIZE).
				ContextMenu(hCtxMenu).
				CtrlExStyles(co.LVS_EX_FULLROWSELECT).
				CtrlStyles(co.LVS_REPORT|co.LVS_NOSORTHEADER|
					co.LVS_SHOWSELALWAYS|co.LVS_SORTASCENDING),
		),
		lstValues: ui.NewListView(wnd,
			ui.ListViewOpts().
				Position(win.POINT{X: 500, Y: 6}).
				Size(win.SIZE{Cx: 242, Cy: 306}).
				Horz(ui.HORZ_REPOS).
				Vert(ui.VERT_RESIZE).
				CtrlExStyles(co.LVS_EX_GRIDLINES).
				CtrlStyles(co.LVS_REPORT|co.LVS_NOSORTHEADER),
		),
		cachedTags: make(map[string]*id3v2.Tag),
	}

	me.eventsMain()
	me.eventsLstFiles()
	me.eventsMenu()
	return me
}

func (me *DlgMain) Run() int {
	defer me.lstFiles.ContextMenu().DestroyMenu()

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
			me.lstFiles.SetRedraw(false)
			for _, mp3 := range mp3s {
				tag := me.cachedTags[mp3]

				padding := "N/A"
				if !tag.IsEmpty() {
					padding = strconv.Itoa(tag.OriginalPadding())
				}

				if item, found := me.lstFiles.Items().Find(mp3); !found { // file not added yet?
					me.lstFiles.Items().AddWithIcon(0, mp3, padding)
				} else {
					item.SetText(1, padding) // update padding
				}
			}
			me.lstFiles.SetRedraw(true)
			me.lstFiles.Columns().SetWidthToFill(0)

			if onFinish != nil {
				onFinish()
			}
		})
	}()
}

func (me *DlgMain) displayFramesOfSelectedFiles() {
	me.lstValues.SetRedraw(false)
	me.lstValues.Items().DeleteAll() // clear all tag displays

	selMp3s := me.lstFiles.Columns().SelectedTexts(0)

	if len(selMp3s) > 1 { // multiple files selected, no tags are shown
		me.lstValues.Items().
			Add("", fmt.Sprintf("%d selected...", len(selMp3s)))

	} else if len(selMp3s) == 1 { // only 1 file selected, we display its tag
		cachedTag := me.cachedTags[selMp3s[0]]

		for _, frameDyn := range cachedTag.Frames() { // read each frame of the tag
			newItem := me.lstValues.Items().
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
					me.lstValues.Items().Add("", (*myFrame.Texts())[i]) // subsequent
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

	me.lstValues.SetRedraw(true)
	me.lstValues.Columns().SetWidthToFill(1)
	me.lstValues.Hwnd().EnableWindow(len(selMp3s) > 0) // if no files selected, disable lstValues
}

func (me *DlgMain) reSaveTagsOfSelectedFiles(onFinish func()) {
	selMp3s := me.lstFiles.Columns().SelectedTexts(0)

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
		me.wnd.Hwnd().SetWindowText(ids.APP_TITLE)
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%s (%d/%d)",
			ids.APP_TITLE, me.lstFiles.Items().SelectedCount(), total))
	}
}
