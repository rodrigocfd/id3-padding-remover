//go:build windows

package dlgpicture

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/ole"
	"github.com/rodrigocfd/windigo/win/ole/oleaut"
)

// Child window to render pictures.
type DlgPicture struct {
	wnd *ui.Control

	rel    *ole.Releaser
	picOle *oleaut.IPicture
}

// Constructor.
func New(parent ui.Parent, x, y, cx, cy int) *DlgPicture {
	me := &DlgPicture{
		wnd: ui.NewControl(parent, ui.OptsControl().
			Position(x, y).
			Size(cx, cy),
		),
		rel:    ole.NewReleaser(),
		picOle: nil,
	}
	me.events()
	return me
}

func (me *DlgPicture) Repaint() error {
	return me.wnd.Hwnd().InvalidateRect(nil, true)
}
