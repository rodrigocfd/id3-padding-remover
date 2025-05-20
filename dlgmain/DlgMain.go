//go:build windows

package dlgmain

import (
	"id3fit/ids"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/ole"
)

// Main application dialog.
type DlgMain struct {
	wnd      *ui.Main
	lstFiles *ui.ListView
	sortCol  int
	sortAsc  bool

	rel        *ole.Releaser
	dropTarget *ole.IDropTarget
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

	rel := ole.NewReleaser()
	dropTarget := ole.NewIDropTargetImpl(rel)

	me := &DlgMain{wnd, lstFiles, sortCol, sortAsc, rel, dropTarget}
	me.events()
	return me
}

func (me *DlgMain) Run() int {
	defer me.rel.Release()
	return me.wnd.RunAsMain()
}
