package main

import (
	"id3fit/id3"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func main() {
	NewDlgMain().Run()
}

type DlgMain struct {
	wnd               ui.WindowMain
	hImageList        win.HIMAGELIST // for system MP3 icon
	lstFiles          ui.ListView
	lstFilesSelLocked bool // LVN_ITEMCHANGED is scheduled to fire
	lstValues         ui.ListView
	resizer           ui.Resizer
	statusBar         ui.StatusBar
	cachedTags        map[string]*id3.Tag // for each file currently in the list
}

func NewDlgMain() *DlgMain {
	wnd := ui.NewWindowMainRaw(&ui.WindowMainRawOpts{
		Title: "ID3 Fit",
		Styles: co.WS_CAPTION | co.WS_SYSMENU | co.WS_CLIPCHILDREN |
			co.WS_BORDER | co.WS_VISIBLE | co.WS_MINIMIZEBOX |
			co.WS_MAXIMIZEBOX | co.WS_SIZEBOX,
		ExStyles:       co.WS_EX_ACCEPTFILES,
		ClientAreaSize: win.SIZE{Cx: 700, Cy: 380},
		IconId:         101,
		AccelTable:     createAccelTable(),
	})

	me := DlgMain{
		wnd:        wnd,
		hImageList: win.ImageListCreate(16, 16, co.ILC_COLOR32, 1, 1),
		lstFiles: ui.NewListViewRaw(wnd, &ui.ListViewRawOpts{
			Position:         win.POINT{X: 6, Y: 6},
			Size:             win.SIZE{Cx: 438, Cy: 346},
			ListViewStyles:   co.LVS_REPORT | co.LVS_NOSORTHEADER | co.LVS_SHOWSELALWAYS,
			ListViewExStyles: co.LVS_EX_FULLROWSELECT,
			ContextMenu:      createContextMenu(),
		}),
		lstValues: ui.NewListViewRaw(wnd, &ui.ListViewRawOpts{
			Position:         win.POINT{X: 450, Y: 6},
			Size:             win.SIZE{Cx: 342, Cy: 346},
			ListViewStyles:   co.LVS_REPORT | co.LVS_NOSORTHEADER,
			ListViewExStyles: co.LVS_EX_GRIDLINES,
		}),
		resizer:    ui.NewResizer(wnd),
		statusBar:  ui.NewStatusBar(wnd),
		cachedTags: make(map[string]*id3.Tag),
	}

	me.eventsMain()
	me.eventsLstFiles()
	me.eventsLstFilesMenu()
	return &me
}

func (me *DlgMain) Run() int {
	me.hImageList.AddIconFromShell("mp3")
	defer me.hImageList.Destroy()

	return me.wnd.RunAsMain()
}
