//go:build windows

package dlgprogress

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/ids"
	"time"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

const _STEP_MS = 50

// Modal dialog to load ID3v2 tags from MP3 files.
type DlgProgressLoad struct {
	wnd        *ui.Modal
	prog       *ui.ProgressBar
	paths      []string
	openedTags []TagAndPath
}

type TagAndPath struct {
	Tag  *id3v2.Tag
	Path string
}

// Constructor; blocks until the modal is closed.
func ShowModalLoad(parent ui.Parent, incomingPaths []string) []TagAndPath {
	wnd := ui.NewModalDlg(parent, ids.DLG_PROGRESS)
	prog := ui.NewProgressBarDlg(wnd, ids.PRO_PRO, ui.LAY_NONE_NONE)

	me := &DlgProgressLoad{wnd, prog, incomingPaths, []TagAndPath{}}
	me.events()
	me.wnd.ShowModal() // blocks until the modal is closed
	return me.openedTags
}

func (me *DlgProgressLoad) events() {
	me.wnd.On().WmInitDialog(func(_ ui.WmInitDialog) bool {
		me.prog.SetMarquee(true)
		me.wnd.Hwnd().SetWindowText("Loading...")
		go me.loadAsync()
		return true
	})
}

func (me *DlgProgressLoad) loadAsync() {
	allPaths := make([]string, 0, len(me.paths)) // grab all files within all subfolders
	for _, incomingPath := range me.paths {
		if win.PathIsFolder(incomingPath) {
			nested, _ := win.PathEnumDeep(incomingPath, "mp3")
			allPaths = append(allPaths, nested...)
		} else if win.PathHasExtension(incomingPath, "mp3") {
			allPaths = append(allPaths, incomingPath)
		}
	}
	if len(allPaths) == 0 {
		me.wnd.UiThread(func() {
			ui.MsgWarn(me.wnd, "No MP3s", "",
				fmt.Sprintf("No MP3s found in %d item(s).", len(me.paths)))
			me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
		})
		return // nothing do to
	}

	me.wnd.UiThread(func() {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("0/%d file(s) read...", len(allPaths)))
		me.prog.SetRange(0, len(allPaths))
		me.prog.SetPos(0)
	})

	me.openedTags = make([]TagAndPath, 0, len(allPaths)) // load and cache all the MP3 tags
	for idxMp3, path := range allPaths {
		pTag, err := id3v2.TagFromFile(path)
		if err != nil {
			me.wnd.UiThread(func() {
				me.wnd.Hwnd().MessageBox(
					fmt.Sprintf("Error loading tag:\n%s\n\n%s", path, err.Error()),
					"Error", co.MB_ICONERROR)
				me.openedTags = []TagAndPath{}
				me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
			})
			return // stop on first error, no tag is loaded
		}
		me.openedTags = append(me.openedTags, TagAndPath{pTag, path})

		me.wnd.UiThread(func() { // UI progress feedback
			me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%d/%d file(s) read...", idxMp3+1, len(allPaths)))
			me.prog.SetPos(idxMp3 + 1)
		})
		win.Sleep(_STEP_MS * time.Millisecond)
	}
	me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
}
