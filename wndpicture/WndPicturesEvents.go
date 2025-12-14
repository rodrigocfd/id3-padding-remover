//go:build windows

package wndpicture

import (
	"github.com/rodrigocfd/windigo/win"
)

func (me *WndPicture) events() {

	me.wnd.On().WmPaint(func() {
		var ps win.PAINTSTRUCT
		hdc, _ := me.wnd.Hwnd().BeginPaint(&ps)
		defer me.wnd.Hwnd().EndPaint(&ps)

		if me.picOle != nil {
			sz, _ := me.picOle.Size()
			me.picOle.Render(hdc,
				win.POINT{},
				win.SIZE{Cx: ps.RcPaint.Right, Cy: ps.RcPaint.Bottom},
				win.POINT{X: 0, Y: sz.Cy},
				win.SIZE{Cx: sz.Cx, Cy: -sz.Cy},
			)
		}
	})

	// me.wnd.On().WmRButtonUp(func(p ui.WmMouse) {
	// 	hSubMenu0, _ := me.hMenu.GetSubMenu(0)
	// 	hSubMenu0.ShowAtPoint(p.Pos(), me.wnd.Hwnd(), me.wnd.Hwnd())
	// })

	// me.wnd.On().WmInitMenuPopup(func(p ui.WmInitMenuPopup) {
	// 	firstId, _ := p.HMenu().GetMenuItemID(0)
	// 	if firstId == ids.MNU_PIC_INSERT {
	// 		hasPic := me.PicOle != nil
	// 		p.HMenu().EnableMenuItemByCmd(hasPic,
	// 			ids.MNU_PIC_EXTRACT, ids.MNU_PIC_DELETE)
	// 	}
	// })

	// me.wnd.On().WmCommandAccelMenu(ids.MNU_PIC_INSERT, func() {
	// 	rel := ole.NewReleaser()
	// 	defer rel.Release()

	// 	fod, _ := ole.CoCreateInstance[shell.IFileOpenDialog](
	// 		rel, co.CLSID_FileOpenDialog, co.CLSCTX_INPROC_SERVER)

	// 	defOpts, _ := fod.GetOptions()
	// 	fod.SetOptions(defOpts |
	// 		co.FOS_FORCEFILESYSTEM |
	// 		co.FOS_FILEMUSTEXIST,
	// 	)

	// 	fod.SetFileTypes([]shell.COMDLG_FILTERSPEC{
	// 		{Name: "BMP files", Spec: "*.bmp"},
	// 		{Name: "JPG files", Spec: "*.jpg"},
	// 		{Name: "PNG files", Spec: "*.png"},
	// 		{Name: "All files", Spec: "*.*"},
	// 	})
	// 	fod.SetFileTypeIndex(2)

	// 	if ok, _ := fod.Show(me.wnd.Hwnd()); ok {
	// 		item, _ := fod.GetResult(rel)
	// 		path, _ := item.GetDisplayName(co.SIGDN_FILESYSPATH)
	// 		me.OnInsert(path)
	// 	}
	// })

	// me.wnd.On().WmCommandAccelMenu(ids.MNU_PIC_EXTRACT, func() {
	// 	rel := ole.NewReleaser()
	// 	defer rel.Release()

	// 	fsd, _ := ole.CoCreateInstance[shell.IFileSaveDialog](
	// 		rel, co.CLSID_FileSaveDialog, co.CLSCTX_INPROC_SERVER)

	// 	fsd.SetFileTypes([]shell.COMDLG_FILTERSPEC{
	// 		{Name: "All files", Spec: "*.*"},
	// 	})
	// 	fsd.SetFileTypeIndex(1)
	// 	fsd.SetFileName("album-cover.jpg")

	// 	if ok, _ := fsd.Show(me.wnd.Hwnd()); ok {
	// 		item, _ := fsd.GetResult(rel)
	// 		destPath, _ := item.GetDisplayName(co.SIGDN_FILESYSPATH)
	// 		me.OnExtract(destPath)
	// 	}
	// })

	// me.wnd.On().WmCommandAccelMenu(ids.MNU_PIC_DELETE, func() {
	// 	resp := ui.MsgOkCancel(parent, "Delete picture", "",
	// 		"Do you want to delete the picture from this tag?", "&Delete")
	// 	if resp == co.ID_OK {
	// 		me.OnDelete()
	// 	}
	// })

}
