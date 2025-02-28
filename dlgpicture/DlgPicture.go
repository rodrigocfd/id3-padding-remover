//go:build windows

package dlgpicture

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/ole"
)

// Child window to render pictures.
type DlgPicture struct {
	wnd *ui.Control
	pic *ole.IPicture
}

// Constructor.
func New(parent ui.Parent, x, y, cx, cy int, pic *ole.IPicture) *DlgPicture {
	me := &DlgPicture{
		wnd: ui.NewControl(parent, ui.Opts.Control().
			Position(x, y).
			Size(cx, cy),
		),
		pic: pic,
	}
	me.events()
	return me
}
