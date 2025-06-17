//go:build windows

package wndpicture

import (
	"id3fit/id3v2"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/ole"
	"github.com/rodrigocfd/windigo/win/ole/oleaut"
)

func (me *WndPicture) LoadPicOle(tags []*id3v2.Tag) {
	apic := id3v2.SameFrameAcrossAllTags("APIC", tags)
	if apic == nil {
		return // we don't have a picture to display
	}

	body, ok := apic.Body().(*id3v2.BodyPicture)
	if !ok {
		ui.MsgError(me.wnd.Parent(), "Picture parsing", "",
			"APIC frame does not contain BodyPicture body type") // should never happen
		return
	}

	localRel := ole.NewReleaser()
	defer localRel.Release()
	stream, err := ole.SHCreateMemStream(localRel, body.Bin)
	if err != nil {
		ui.MsgError(me.wnd.Parent(), "Picture stream", "",
			"Failed to stream picture:\n"+err.Error())
		return
	}

	me.rel.ReleaseNow(me.picOle) // free right away, before setting new IPicture
	me.picOle, err = oleaut.OleLoadPicture(me.rel, stream, uint(len(body.Bin)), true)
	if err != nil {
		ui.MsgError(me.wnd.Parent(), "Picture loading", "",
			"Failed to load picture:\n"+err.Error())
	}
}

func (me *WndPicture) PicOle() *oleaut.IPicture {
	return me.picOle
}
