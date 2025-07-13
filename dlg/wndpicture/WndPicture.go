//go:build windows

package wndpicture

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

// Child window to render pictures.
type WndPicture struct {
	wnd *ui.Control

	rel         *win.OleReleaser
	picOle      *win.IPicture
	picNumBytes uint
}

// Constructor.
func New(parent ui.Parent, x, y, cx, cy int, rel *win.OleReleaser) *WndPicture {
	me := &WndPicture{
		wnd: ui.NewControl(parent,
			ui.OptsControl().
				Position(x, y).
				Size(cx, cy),
		),
		rel:    rel,
		picOle: nil,
	}
	me.events()
	return me
}

func (me *WndPicture) Repaint() error {
	return me.wnd.Hwnd().InvalidateRect(nil, true)
}
