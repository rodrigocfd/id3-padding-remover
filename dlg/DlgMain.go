//go:build windows

package dlg

import (
	"github.com/rodrigocfd/windigo/ui"
)

// Main application dialog.
type DlgMain struct {
	wnd      *ui.Main
	lstFiles *ui.ListView

	sortCol int
	sortAsc bool
}

// Constructor; blocks until the window is closed.
func RunMain() int {
	wnd := ui.NewMainDlg(
		ui.OptsMainDlg().
			DlgId(DLG_MAIN).
			IconId(ICO_MAIN).
			AccelTableId(ACC_MAIN).
			DropFiles(true),
	)
	lstFiles := ui.NewListViewDlg(wnd, LST_FILES, MNU_FILE, ui.LAY_RESIZE_RESIZE)
	sortCol := 0
	sortAsc := true

	me := &DlgMain{wnd, lstFiles, sortCol, sortAsc}
	me.events()
	return me.wnd.RunAsMain()
}
