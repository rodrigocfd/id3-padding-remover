//go:build windows

package dlgmain

import (
	"id3fit/dlg/ids"

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
}

// Constructor; blocks until the window is closed.
func RunNew() int {
	wnd := ui.NewMainDlg(
		ui.OptsMainDlg().
			DlgId(ids.DLG_MAIN).
			IconId(ids.ICO_MAIN).
			AccelTableId(ids.ACC_MAIN),
	)
	lstFiles := ui.NewListViewDlg(wnd, ids.LST_FILES, ids.MNU_FILE, ui.LAY_RESIZE_RESIZE)
	sortCol := 0
	sortAsc := true

	rel := win.NewOleReleaser()
	defer rel.Release()
	dropTarget := win.NewIDropTargetImpl(rel)

	me := &DlgMain{wnd, lstFiles, sortCol, sortAsc, dropTarget}
	me.events()
	return me.wnd.RunAsMain()
}
