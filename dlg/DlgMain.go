//go:build windows

package dlg

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

// Main application dialog.
type DlgMain struct {
	wnd      *ui.Main
	lstFiles *ui.ListView

	dropTarget *win.IDropTarget
	sortCol    int
	sortAsc    bool
}

// Constructor; blocks until the window is closed.
func RunMain() int {
	rel := win.NewOleReleaser()
	defer rel.Release()

	wnd := ui.NewMainDlg(
		ui.OptsMainDlg().
			DlgId(DLG_MAIN).
			IconId(ICO_MAIN).
			AccelTableId(ACC_MAIN),
	)
	lstFiles := ui.NewListViewDlg(wnd, LST_FILES, MNU_FILE, ui.LAY_RESIZE_RESIZE)
	dropTarget := win.NewIDropTargetImpl(rel)
	sortCol := 0
	sortAsc := true

	me := &DlgMain{wnd, lstFiles, dropTarget, sortCol, sortAsc}
	me.events()
	return me.wnd.RunAsMain()
}
