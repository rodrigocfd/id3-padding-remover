package main

import (
	"id3fit/dlgfields"
	"id3fit/id3v2"
	"runtime"

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
		dlgFields:  dlgfields.NewDlgFields(wnd, win.POINT{X: 292, Y: 4}, ui.HORZ_REPOS, ui.VERT_NONE),
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
