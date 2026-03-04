//go:build windows

package dlg

import (
	"fmt"
	"runtime"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/wstr"
)

func (me *DlgMain) events() {

	me.wnd.On().WmInitDialog(func(_ ui.WmInitDialog) bool {
		me.lstFiles.SetExtendedStyle(true, co.LVS_EX_FULLROWSELECT)

		me.lstFiles.AddCol("File", ui.DpiX(400)).SetSortArrow(co.HDF_SORTUP)
		me.lstFiles.AddCol("Pad", ui.DpiX(50)).SetJustification(co.HDF_RIGHT)
		me.lstFiles.AddCol("Pic", ui.DpiX(30)).SetJustification(co.HDF_CENTER)
		me.lstFiles.AddCol("RG", ui.DpiX(30)).SetJustification(co.HDF_CENTER)
		me.lstFiles.AddCol("Artist", ui.DpiX(90))
		me.lstFiles.AddCol("Year", ui.DpiX(40)).SetJustification(co.HDF_CENTER)
		me.lstFiles.AddCol("Album", ui.DpiX(100))
		me.lstFiles.AddCol("T#", ui.DpiX(30)).SetJustification(co.HDF_RIGHT)
		me.lstFiles.AddCol("Title", ui.DpiX(100))
		me.lstFiles.AddCol("Genre", ui.DpiX(90))
		me.lstFiles.AddCol("Performer", ui.DpiX(70))
		me.lstFiles.AddCol("Composer", ui.DpiX(70))
		me.lstFiles.AddCol("Lyricist", ui.DpiX(70))
		me.lstFiles.AddCol("Orig. artist", ui.DpiX(70))
		me.lstFiles.AddCol("Comment", ui.DpiX(70))

		me.lstFiles.Col(0).SetWidthToFill()
		return true
	})

	me.wnd.On().WmSize(func(p ui.WmSize) {
		if p.Request() != co.SIZE_REQ_MINIMIZED {
			me.lstFiles.Col(0).SetWidthToFill()
		}
	})

	me.wnd.On().WmInitMenuPopup(func(p ui.WmInitMenuPopup) {
		firstId, _ := p.HMenu().GetMenuItemID(0)
		if firstId == MNU_FILE_OPEN {
			p.HMenu().SetMenuDefaultItemByCmd(MNU_FILE_EDIT)

			enable := me.lstFiles.SelectedItemCount() > 0
			p.HMenu().EnableMenuItemByCmd(enable,
				MNU_FILE_EDIT,
				MNU_FILE_RESAVE,
				MNU_FILE_REMOVE,
				MNU_FILE_DELPIC,
				MNU_FILE_DELPICRG)
		}
	})

	me.wnd.On().WmDropFiles(func(p ui.WmDropFiles) {
		paths, _ := p.HDrop().DragQueryFile()
		me.addMp3sToList(paths)
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FILE_OPEN, func() {
		oleRel := win.NewOleReleaser()
		defer oleRel.Release()

		var fod *win.IFileOpenDialog
		win.CoCreateInstance(oleRel, &co.CLSID_FileOpenDialog, nil, co.CLSCTX_INPROC_SERVER, &fod)

		defOpts, _ := fod.GetOptions()
		fod.SetOptions(defOpts |
			co.FOS_FORCEFILESYSTEM |
			co.FOS_FILEMUSTEXIST |
			co.FOS_ALLOWMULTISELECT,
		)

		fod.SetFileTypes([]win.COMDLG_FILTERSPEC{
			{Name: "MP3 files", Spec: "*.mp3"},
			{Name: "All files", Spec: "*.*"},
		})
		fod.SetFileTypeIndex(1)

		if ok, _ := fod.Show(me.wnd.Hwnd()); ok {
			arr, _ := fod.GetResults(oleRel)
			paths, _ := arr.EnumDisplayNames(co.SIGDN_FILESYSPATH)
			me.addMp3sToList(paths)
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FILE_EDIT, func() {
		if me.editSelected() {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FILE_REMOVE, func() {
		me.lstFiles.DeleteSelectedItems()
		me.updateTitlebarCount()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FILE_RESAVE, func() {
		text := fmt.Sprintf("Do you want to rewrite the tags of %d file(s)?",
			me.lstFiles.SelectedItemCount())
		if ui.MsgOkCancel(me.wnd, "Save files", "", text, "&Save") {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FILE_DELPIC, func() {
		if me.removePicRg(false) {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FILE_DELPICRG, func() {
		if me.removePicRg(true) {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FILE_ABOUT, func() {
		hInst, _ := win.GetModuleHandle("")
		exeName, _ := hInst.GetModuleFileName()
		nfo, _ := win.VersionLoad(exeName)

		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)

		caption := fmt.Sprintf("%s %d.%d.%d",
			nfo.ProductName, nfo.Version[0], nfo.Version[1], nfo.Version[2])

		text := fmt.Sprintf(
			"%s\n\n"+
				"Compiler: %s\n"+
				"GC cycles: %d\n"+
				"Alloc: %s\n"+
				"Next GC: %s\n"+
				"Frees: %d",
			nfo.LegalCopyright,
			runtime.Version(),
			stats.NumGC, wstr.FmtBytes(int(stats.HeapAlloc)),
			wstr.FmtBytes(int(stats.NextGC)), stats.Frees)

		ui.MsgOk(me.wnd, "About", caption, text)
	})

	me.lstFiles.On().NmDblClk(func(_ *win.NMITEMACTIVATE) {
		if me.editSelected() {
			me.saveSelected()
		}
	})

	me.lstFiles.On().LvnKeyDown(func(p *win.NMLVKEYDOWN) {
		switch p.WVKey {
		case co.VK_DELETE:
			me.lstFiles.DeleteSelectedItems()
			me.updateTitlebarCount()
		case co.VK_RETURN: // Enter key
			if me.editSelected() {
				me.saveSelected()
			}
		}
	})

	me.lstFiles.On().LvnItemChanged(func(_ *win.NMLISTVIEW) {
		me.updateTitlebarCount()
	})

	me.lstFiles.Header().On().HdnItemClick(func(p *win.NMHEADER) {
		lvCol := me.lstFiles.Col(int(p.IItem))
		if lvCol.Index() != me.sortCol { // user changed the column
			lvCol.SetSortArrow(co.HDF_SORTUP)
			me.sortAsc = true
		} else { // user is reversing the same column
			curArrow := lvCol.SortArrow()
			if curArrow == co.HDF_SORTUP {
				lvCol.SetSortArrow(co.HDF_SORTDOWN)
				me.sortAsc = false
			} else {
				lvCol.SetSortArrow(co.HDF_SORTUP)
				me.sortAsc = true
			}
		}
		me.sortCol = lvCol.Index()
		me.sortList()
	})

}
