//go:build windows

package dlgmain

import (
	"id3fit/ids"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

// Main application dialog.
type DlgMain struct {
	wnd      *ui.Main
	lstFiles *ui.ListView

	sortCol    int
	sortAsc    bool
	dropTarget *win.IDropTarget

	hCursorWait win.HCURSOR
	isWaiting   bool // Set by setWaitState(), read at WM_SETCURSOR.
}

// Constructor; blocks until the window is closed.
func RunMain() int {
	wnd := ui.NewMainDlg(
		ui.OptsMainDlg().
			DlgId(ids.DLG_MAIN).
			IconId(ids.ICO_MAIN).
			AccelTableId(ids.ACC_MAIN),
	)
	lstFiles := ui.NewListViewDlg(wnd, ids.LST_FILES, ids.MNU_FILE, ui.LAY_RESIZE_RESIZE)
	sortCol := 0
	sortAsc := true

	oleRel := win.NewOleReleaser()
	defer oleRel.Release()
	dropTarget := win.NewIDropTargetImpl(oleRel)

	hCursorWait, _ := win.HINSTANCE(0).LoadCursor(win.CursorResIdc(co.IDC_WAIT))
	isWaiting := false

	me := &DlgMain{wnd, lstFiles,
		sortCol, sortAsc, dropTarget,
		hCursorWait, isWaiting}
	me.events()
	return me.wnd.RunAsMain()
}
