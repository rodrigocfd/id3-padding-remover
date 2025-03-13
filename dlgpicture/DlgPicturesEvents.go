//go:build windows

package dlgpicture

import (
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
)

func (me *DlgPicture) events() {

	me.wnd.On().WmPaint(func() {
		var ps win.PAINTSTRUCT
		hdc, _ := me.wnd.Hwnd().BeginPaint(&ps)
		defer me.wnd.Hwnd().EndPaint(&ps)

		if me.pic.Ppvt() != nil {
			hdcScreen, _ := win.HWND(0).GetDC()
			defer win.HWND(0).ReleaseDC(hdcScreen)

			cxhm, _ := me.pic.Width()
			cyhm, _ := me.pic.Height()

			me.pic.Render(hdc,
				win.POINT{},
				win.SIZE{Cx: ps.RcPaint.Right, Cy: ps.RcPaint.Bottom},
				win.POINT{X: 0, Y: int32(cyhm)},
				win.SIZE{Cx: int32(cxhm), Cy: int32(-cyhm)},
			)
		}
	})

	me.wnd.On().WmRButtonUp(func(p wm.Mouse) {
		println("oi")
	})

}
