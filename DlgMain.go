package main

import (
	"id3fit/id3"
	"runtime"

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
	resizer           ui.Resizer
	statusBar         ui.StatusBar
	cachedTags        map[string]*id3.Tag // for each file currently in the list
}

func NewDlgMain() *DlgMain {
	wnd := ui.NewWindowMainRaw(ui.WindowMainRawOpts{
		Title:          APP_TITLE,
		ClientAreaSize: win.SIZE{Cx: 750, Cy: 340},
		IconId:         ICO_MAIN,
		AccelTable:     createAccelTable(),
		ExStyles:       co.WS_EX_ACCEPTFILES,
		Styles: co.WS_CAPTION | co.WS_SYSMENU | co.WS_CLIPCHILDREN |
			co.WS_BORDER | co.WS_VISIBLE | co.WS_MINIMIZEBOX |
			co.WS_MAXIMIZEBOX | co.WS_SIZEBOX,
	})

	me := &DlgMain{
		wnd: wnd,
		lstFiles: ui.NewListViewRaw(wnd, ui.ListViewRawOpts{
			Position:         win.POINT{X: 6, Y: 6},
			Size:             win.SIZE{Cx: 488, Cy: 306},
			ContextMenu:      createContextMenu(),
			ListViewExStyles: co.LVS_EX_FULLROWSELECT,
			ListViewStyles: co.LVS_REPORT | co.LVS_NOSORTHEADER |
				co.LVS_SHOWSELALWAYS | co.LVS_SORTASCENDING,
		}),
		lstValues: ui.NewListViewRaw(wnd, ui.ListViewRawOpts{
			Position:         win.POINT{X: 500, Y: 6},
			Size:             win.SIZE{Cx: 242, Cy: 306},
			ListViewExStyles: co.LVS_EX_GRIDLINES,
			ListViewStyles:   co.LVS_REPORT | co.LVS_NOSORTHEADER,
		}),
		resizer:    ui.NewResizer(wnd),
		statusBar:  ui.NewStatusBar(wnd),
		cachedTags: make(map[string]*id3.Tag),
	}

	me.resizer.Add(ui.RESZ_RESIZE, ui.RESZ_RESIZE, me.lstFiles).
		Add(ui.RESZ_REPOS, ui.RESZ_RESIZE, me.lstValues)

	me.eventsMain()
	me.eventsLstFiles()
	me.eventsMenu()
	return me
}

func (me *DlgMain) Run() int {
	defer me.lstFiles.ImageList(co.LVSIL_SMALL).Destroy()
	defer me.lstFiles.ContextMenu().DestroyMenu()

	return me.wnd.RunAsMain()
}
