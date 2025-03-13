//go:build windows

package dlgmain

import (
	"id3fit/ids"

	"github.com/rodrigocfd/windigo/ui"
)

// Main application dialog.
type DlgMain struct {
	wnd      *ui.Main
	lstFiles *ui.ListView
	sortCol  int
	sortAsc  bool
}

// Constructor.
func New() *DlgMain {
	wnd := ui.NewMainDlg(
		ui.Opts.MainDlg().
			DlgId(ids.DLG_MAIN).
			IconId(ids.ICO_MAIN).
			AccelTableId(ids.ACC_MAIN).
			AllowDragDrop(true),
	)
	lstFiles := ui.NewListViewDlg(wnd, ids.LST_FILES, ids.MNU_FILE, ui.LAY_RESIZE_RESIZE)
	sortCol := 0
	sortAsc := true

	me := &DlgMain{wnd, lstFiles, sortCol, sortAsc}
	me.events()
	return me
}

func (me *DlgMain) Run() int {
	return me.wnd.RunAsMain()
}
