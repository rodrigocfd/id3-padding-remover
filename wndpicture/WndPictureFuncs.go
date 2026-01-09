//go:build windows

package wndpicture

import (
	"id3fit/id3v2"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

// Loads the cover art, if due, into the IPicture COM object.
func (me *WndPicture) LoadPicture(tags []*id3v2.Tag) (pixels win.SIZE, nBytes int) {
	apic := id3v2.SameFrameAcrossAllTags("APIC", tags)
	if apic == nil {
		return win.SIZE{}, 0 // we don't have a picture to display
	}

	body, ok := apic.Body().(*id3v2.BodyPicture)
	if !ok {
		ui.MsgError(me.wnd.Parent(), "Picture parsing", "",
			"APIC frame does not contain BodyPicture body type") // should never happen
		return win.SIZE{}, 0
	}

	localOleRel := win.NewOleReleaser()
	defer localOleRel.Release()

	iStream, err := win.SHCreateMemStream(localOleRel, body.Bin) // create IStream over pic data
	if err != nil {
		ui.MsgError(me.wnd.Parent(), "Picture stream", "",
			"Failed to stream picture:\n"+err.Error())
		return win.SIZE{}, 0
	}

	me.oleRel.ReleaseNow(me.iPic) // free IPicture right away, before loading new

	me.iPic, err = win.OleLoadPicture(me.oleRel, iStream, len(body.Bin), true)
	if err != nil {
		ui.MsgError(me.wnd.Parent(), "Picture loading", "",
			"Failed to load picture:\n"+err.Error())
		return win.SIZE{}, 0
	}

	hdcScreen, _ := win.HWND(0).GetDC()
	defer win.HWND(0).ReleaseDC(hdcScreen)
	szPic, _ := me.iPic.SizePixels(hdcScreen) // picture resolution in pixels

	return szPic, len(body.Bin)
}
