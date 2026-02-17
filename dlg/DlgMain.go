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

	sortCol    int
	sortAsc    bool
	dropTarget *win.IDropTarget
}

// Constructor; blocks until the window is closed.
func RunMain() int {
	wnd := ui.NewMainDlg(
		ui.OptsMainDlg().
			DlgId(DLG_MAIN).
			IconId(ICO_MAIN).
			AccelTableId(ACC_MAIN),
	)
	lstFiles := ui.NewListViewDlg(wnd, LST_FILES, MNU_FILE, ui.LAY_RESIZE_RESIZE)
	sortCol := 0
	sortAsc := true

	oleRel := win.NewOleReleaser()
	defer oleRel.Release()
	dropTarget := win.NewIDropTargetImpl(oleRel)

	me := &DlgMain{wnd, lstFiles, sortCol, sortAsc, dropTarget}
	me.events()
	return me.wnd.RunAsMain()
}
