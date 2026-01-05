//go:build windows

package dlgmain

import (
	"fmt"
	"id3fit/ids"
	"runtime"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/wstr"
)

func (me *DlgMain) events() {

	me.wnd.On().WmInitDialog(func(_ ui.WmInitDialog) bool {
		me.wnd.Hwnd().RegisterDragDrop(me.dropTarget)

		me.lstFiles.ImageList(co.LVSIL_SMALL).AddIconFromShell("mp3")
		me.lstFiles.SetExtendedStyle(true, co.LVS_EX_FULLROWSELECT)

		me.lstFiles.Cols.Add("File", ui.DpiX(400)).SetSortArrow(co.HDF_SORTUP)
		me.lstFiles.Cols.Add("Pad", ui.DpiX(50)).SetJustification(co.HDF_RIGHT)
		me.lstFiles.Cols.Add("Pic", ui.DpiX(30)).SetJustification(co.HDF_CENTER)
		me.lstFiles.Cols.Add("RG", ui.DpiX(30)).SetJustification(co.HDF_CENTER)
		me.lstFiles.Cols.Add("Artist", ui.DpiX(90))
		me.lstFiles.Cols.Add("Year", ui.DpiX(40)).SetJustification(co.HDF_CENTER)
		me.lstFiles.Cols.Add("Album", ui.DpiX(100))
		me.lstFiles.Cols.Add("T#", ui.DpiX(30)).SetJustification(co.HDF_RIGHT)
		me.lstFiles.Cols.Add("Title", ui.DpiX(100))
		me.lstFiles.Cols.Add("Genre", ui.DpiX(90))
		me.lstFiles.Cols.Add("Performer", ui.DpiX(70))
		me.lstFiles.Cols.Add("Composer", ui.DpiX(70))
		me.lstFiles.Cols.Add("Lyricist", ui.DpiX(70))
		me.lstFiles.Cols.Add("Orig. artist", ui.DpiX(70))
		me.lstFiles.Cols.Add("Comment", ui.DpiX(70))

		me.lstFiles.Cols.Get(0).SetWidthToFill()
		return true
	})

	me.wnd.On().WmSize(func(p ui.WmSize) {
		if p.Request() != co.SIZE_REQ_MINIMIZED {
			me.lstFiles.Cols.Get(0).SetWidthToFill()
		}
	})

	me.wnd.On().WmInitMenuPopup(func(p ui.WmInitMenuPopup) {
		firstId, _ := p.HMenu().GetMenuItemID(0)
		if firstId == ids.MNU_FILE_OPEN {
			p.HMenu().SetMenuDefaultItemByCmd(ids.MNU_FILE_EDIT)

			enable := me.lstFiles.Items.SelectedCount() > 0
			p.HMenu().EnableMenuItemByCmd(enable,
				ids.MNU_FILE_EDIT,
				ids.MNU_FILE_RESAVE,
				ids.MNU_FILE_REMOVE,
				ids.MNU_FILE_DELPIC,
				ids.MNU_FILE_DELPICRG)
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_OPEN, func() {
		oleRel := win.NewOleReleaser()
		defer oleRel.Release()

		var fod *win.IFileOpenDialog
		win.CoCreateInstance(oleRel, co.CLSID_FileOpenDialog, nil, co.CLSCTX_INPROC_SERVER, &fod)

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

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_EDIT, func() {
		if me.editSelected() {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_REMOVE, func() {
		me.lstFiles.Items.DeleteSelected()
		me.updateTitlebarCount()
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_RESAVE, func() {
		text := fmt.Sprintf("Do you want to rewrite the tags of %d file(s)?",
			me.lstFiles.Items.SelectedCount())
		if ui.MsgOkCancel(me.wnd, "Save files", "", text, "&Save") == co.ID_OK {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_DELPIC, func() {
		if me.removePicRg(false) {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_DELPICRG, func() {
		if me.removePicRg(true) {
			me.saveSelected()
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_ABOUT, func() {
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
			me.lstFiles.Items.DeleteSelected()
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
		lvCol := me.lstFiles.Cols.Get(int(p.IItem))
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

	me.dropTarget.Drop(
		func(dataObj *win.IDataObject, _ co.MK, _ win.POINT, _ *co.DROPEFFECT) co.HRESULT {
			fetc := win.FORMATETC{
				CfFormat: co.CF_HDROP,
				Aspect:   co.DVASPECT_CONTENT,
				Lindex:   -1,
				Tymed:    co.TYMED_HGLOBAL,
			}

			stg, err := dataObj.GetData(&fetc)
			if err != nil {
				ui.MsgError(me.wnd, "Drop error", "", err.Error())
				return co.HRESULT_S_OK
			}
			defer win.ReleaseStgMedium(&stg)

			if hGlobal, ok := stg.HGlobal(); ok {
				hMem, _ := hGlobal.GlobalLock()
				defer hGlobal.GlobalUnlock()

				hDrop := win.HDROP(hMem) // DragFinish() crashes ReleaseStgMedium(), don't call
				paths, _ := hDrop.DragQueryFile()
				me.addMp3sToList(paths)
			}
			return co.HRESULT_S_OK
		},
	)

}
