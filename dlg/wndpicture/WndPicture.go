//go:build windows

package wndpicture

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/ole"
	"github.com/rodrigocfd/windigo/win/ole/oleaut"
)

// Child window to render pictures.
type WndPicture struct {
	wnd *ui.Control

	rel    *ole.Releaser
	picOle *oleaut.IPicture
}

// Constructor.
func New(parent ui.Parent, x, y, cx, cy int) *WndPicture {
	me := &WndPicture{
		wnd: ui.NewControl(parent,
			ui.OptsControl().
				Position(x, y).
				Size(cx, cy),
		),
		rel:    ole.NewReleaser(), // released in WM_DESTROY
		picOle: nil,
	}
	me.events()
	return me
}

func (me *WndPicture) Repaint() error {
	return me.wnd.Hwnd().InvalidateRect(nil, true)
}
