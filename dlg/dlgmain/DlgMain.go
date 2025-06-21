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
	sortCol  int
	sortAsc  bool

	rel        *win.OleReleaser
	dropTarget *win.IDropTarget
}

// Constructor.
func New() *DlgMain {
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
	dropTarget := win.NewIDropTargetImpl(rel)

	me := &DlgMain{wnd, lstFiles, sortCol, sortAsc, rel, dropTarget}
	me.events()
	return me
}

func (me *DlgMain) Run() int {
	defer me.rel.Release() // COM objects cleanup
	return me.wnd.RunAsMain()
}
