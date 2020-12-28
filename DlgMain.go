package main

import (
	"id3-fit/id3"
	"windigo/co"
	"windigo/ui"
)

func main() {
	NewDlgMain().Run()
}

type DlgMain struct {
	wnd               *ui.WindowMain
	iconImgList       *ui.ImageList // for system MP3 icon
	lstFiles          *ui.ListView
	lstFilesMenu      *ui.Menu // files list view right-click menu
	lstFilesSelLocked bool     // LVN_ITEMCHANGED is scheduled to fire
	lstValues         *ui.ListView
	resizer           *ui.Resizer
	statusBar         *ui.StatusBar
	cachedTags        map[string]*id3.Tag // for each file currently in the list
}

func NewDlgMain() *DlgMain {
	wnd := ui.NewWindowMain(
		&ui.WindowMainOpts{
			Title:          "ID3 Fit",
			StylesAdd:      co.WS_MINIMIZEBOX | co.WS_MAXIMIZEBOX | co.WS_SIZEBOX,
			ExStylesAdd:    co.WS_EX_ACCEPTFILES,
			ClientAreaSize: ui.Size{Cx: 700, Cy: 380},
			IconId:         101,
		},
	)

	me := DlgMain{
		wnd:          wnd,
		iconImgList:  ui.NewImageList(16, 16),
		lstFiles:     ui.NewListView(wnd),
		lstFilesMenu: ui.NewMenu(),
		lstValues:    ui.NewListView(wnd),
		resizer:      ui.NewResizer(wnd),
		statusBar:    ui.NewStatusBar(wnd),
		cachedTags:   make(map[string]*id3.Tag),
	}

	me.eventsMain()
	me.eventsLstFiles()
	me.eventsLstFilesMenu()
	return &me
}

func (me *DlgMain) Run() int {
	me.iconImgList.AddShellIcon("mp3")
	defer me.iconImgList.Destroy()

	me.buildLstFilesMenuAndAccel()
	defer me.lstFilesMenu.Destroy()

	return me.wnd.RunAsMain()
}
