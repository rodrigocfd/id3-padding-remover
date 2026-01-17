//go:build windows

package wndpicture

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

// Child window to render pictures.
type WndPicture struct {
	wnd      *ui.Control
	HBmp     win.HBITMAP // deleted by parent window
	szPixels win.SIZE
}

// Constructor.
func New(parent ui.Parent, x, y, cx, cy int) *WndPicture {
	me := &WndPicture{
		wnd: ui.NewControl(parent,
			ui.OptsControl().
				Position(x, y).
				Size(cx, cy),
		),
		HBmp:     win.HBITMAP(0),
		szPixels: win.SIZE{},
	}
	me.events()
	return me
}
