package main

import (
	"fmt"
	"id3fit/id3"
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
	cachedTags        map[string]*id3.Tag // for each file currently in the list
}

func NewDlgMain() *DlgMain {
	hAccel, hCtxMenu := createAccelTableAndMenu()

	wnd := ui.NewWindowMain(
		ui.WindowMainOpts().
			Title(APP_TITLE).
			ClientArea(win.SIZE{Cx: 750, Cy: 320}).
			IconId(ICO_MAIN).
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
		cachedTags: make(map[string]*id3.Tag),
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

// func (me *DlgMain) addFilesToList(mp3s []string, onFinish func()) {
// 	go func() {
// 		type OutputData struct {
// 			Mp3 string
// 			Err error
// 			Tag *id3.Tag
// 		}

// 		// Parse all tags from files in parallel.

// 		processChan := make(chan OutputData, len(mp3s))
// 		outputUnits := make([]OutputData, 0, len(mp3s)) // will receive processing results

// 		for _, mp3 := range mp3s {
// 			go func(mp3 string) {
// 				tag, err := id3.ReadTagFromFile(mp3)
// 				processChan <- OutputData{ // send all results to channel
// 					Mp3: mp3,
// 					Err: err,
// 					Tag: tag,
// 				}
// 			}(mp3)
// 		}
// 		for i := 0; i < len(mp3s); i++ {
// 			outputUnits = append(outputUnits, <-processChan) // receive all results from channel
// 		}

// 		// Back to UI thread, display results.

// 		me.wnd.RunUiThread(func() {
// 			for _, resu := range outputUnits {
// 				if resu.Err != nil {
// 					prompt.Error(me.wnd, "Error parsing tag", nil,
// 						fmt.Sprintf("File:\n%s\n\n%s", resu.Mp3, resu.Err))
// 				} else {
// 					if item, found := me.lstFiles.Items().Find(resu.Mp3); !found { // file not added yet?
// 						me.lstFiles.Items().
// 							AddWithIcon(0, resu.Mp3, strconv.Itoa(resu.Tag.OriginalPadding())) // will fire LVN_INSERTITEM
// 					} else {
// 						item.SetText(1, strconv.Itoa(resu.Tag.OriginalPadding())) // update padding info
// 					}
// 					me.cachedTags[resu.Mp3] = resu.Tag // cache (or re-cache) the tag
// 				}
// 			}
// 			me.lstFiles.Columns().SetWidthToFill(0)
// 			me.displayFramesOfSelectedFiles()
// 			if onFinish != nil {
// 				onFinish()
// 			}
// 		})
// 	}()
// }

func (me *DlgMain) addFilesToList(mp3s []string, onFinish func()) {
	go func() { // launch a separated thread
		halted := false

		for _, mp3 := range mp3s {
			tag, err := id3.ReadTagFromFile(mp3) // read all files sequentially
			if err != nil {
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
				if item, found := me.lstFiles.Items().Find(mp3); !found { // file not added yet?
					me.lstFiles.Items().
						AddWithIcon(0, mp3, strconv.Itoa(tag.OriginalPadding()))
				} else {
					item.SetText(1, strconv.Itoa(tag.OriginalPadding())) // update padding
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

// func (me *DlgMain) reSaveTagsOfSelectedFiles(onFinish func()) {
// 	type InputData struct {
// 		Mp3 string
// 		Tag *id3.Tag
// 	}
// 	type OutputData struct {
// 		Mp3 string
// 		Err error
// 	}

// 	inputUnits := make([]InputData, 0, me.lstFiles.Items().SelectedCount())
// 	for _, selItem := range me.lstFiles.Items().Selected() {
// 		selMp3 := selItem.Text(0)
// 		inputUnits = append(inputUnits, InputData{ // prepare data to be worked upon
// 			Mp3: selMp3,
// 			Tag: me.cachedTags[selMp3],
// 		})
// 	}

// 	go func() {
// 		processChan := make(chan OutputData, len(inputUnits))
// 		outputUnits := make([]OutputData, 0, len(inputUnits)) // will receive processing results

// 		for i := range inputUnits {
// 			go func(i int) {
// 				selUnit := inputUnits[i]
// 				err := selUnit.Tag.SerializeToFile(selUnit.Mp3)
// 				processChan <- OutputData{ // send all results in parallel
// 					Mp3: selUnit.Mp3,
// 					Err: err,
// 				}
// 			}(i)
// 		}
// 		for i := 0; i < len(inputUnits); i++ {
// 			outputUnits = append(outputUnits, <-processChan) // receive all results
// 		}

// 		me.wnd.RunUiThread(func() {
// 			reCachedMp3s := make([]string, 0, len(outputUnits))

// 			for _, outputUnit := range outputUnits { // analyze all results
// 				if outputUnit.Err != nil {
// 					prompt.Error(me.wnd, "Writing error", nil,
// 						fmt.Sprintf("Failed to write tag to:\n%s\n\n%s",
// 							outputUnit.Mp3, outputUnit.Err.Error()))
// 				} else {
// 					reCachedMp3s = append(reCachedMp3s, outputUnit.Mp3)
// 				}
// 			}
// 			me.addFilesToList(reCachedMp3s, onFinish)
// 		})
// 	}()
// }

func (me *DlgMain) reSaveTagsOfSelectedFiles(onFinish func()) {
	go func() { // launch a separated thread
		halted := false
		selMp3s := make([]string, 0, me.lstFiles.Items().SelectedCount())

		for _, selItem := range me.lstFiles.Items().Selected() {
			mp3 := selItem.Text(0)
			selMp3s = append(selMp3s, mp3)
			tag := me.cachedTags[mp3]
			if err := tag.SerializeToFile(mp3); err != nil {
				prompt.Error(me.wnd, "Writing error", nil,
					fmt.Sprintf("Failed to write tag to:\n%sn\n\n%s", mp3, err.Error()))
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
			APP_TITLE, me.lstFiles.Items().SelectedCount(), total))
	}
}

func (me *DlgMain) tellElapsedTime(initCounter int64, numFiles int) {
	freq := float64(win.QueryPerformanceFrequency())
	t0 := float64(initCounter)
	tFinal := float64(win.QueryPerformanceCounter())

	prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
		fmt.Sprintf("%d file(s) processed in %.2f ms.",
			numFiles, ((tFinal-t0)/freq)*1000,
		),
	)
}
