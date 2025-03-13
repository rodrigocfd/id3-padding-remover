//go:build windows

package dlgmain

import (
	"fmt"
	"id3fit/ids"
	"id3fit/slices2"
	"runtime"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/ole"
	"github.com/rodrigocfd/windigo/win/ole/shell"
)

func (me *DlgMain) events() {

	me.wnd.On().WmInitDialog(func(_ wm.InitDialog) bool {
		hImg, _ := win.ImageListCreate(16, 16, co.ILC_COLOR32, 1, 1)
		hImg.AddIconFromShell("mp3")
		me.lstFiles.SetImageList(co.LVSIL_SMALL, hImg) // owned, no co.LVS_SHAREIMAGELISTS

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
		return false
	})

	me.wnd.On().WmSize(func(p wm.Size) {
		if p.Request() != co.SIZE_REQ_MINIMIZED {
			me.lstFiles.Cols.Get(0).SetWidthToFill()
		}
	})

	me.wnd.On().WmInitMenuPopup(func(p wm.InitMenuPopup) {
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

	me.wnd.On().WmDropFiles(func(p wm.DropFiles) {
		paths, _ := slices2.CollectErr(p.HDrop().Iter())
		me.withWaitCursor(func() {
			me.addMp3sToList(paths)
		})
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_OPEN, func() {
		rel := ole.NewReleaser()
		defer rel.Release()

		fod, _ := shell.NewIFileOpenDialog(
			ole.CoCreateInstance(&rel, co.CLSID_FileOpenDialog,
				co.CLSCTX_INPROC_SERVER, co.IID_IFileOpenDialog),
		)

		defOpts, _ := fod.GetOptions()
		fod.SetOptions(defOpts |
			co.FOS_FORCEFILESYSTEM |
			co.FOS_FILEMUSTEXIST |
			co.FOS_ALLOWMULTISELECT,
		)

		fod.SetFileTypes([]shell.COMDLG_FILTERSPEC{
			{Name: "MP3 files", Spec: "*.mp3"},
			{Name: "All files", Spec: "*.*"},
		})
		fod.SetFileTypeIndex(1)

		if ok, _ := fod.Show(me.wnd.Hwnd()); ok {
			arr, _ := fod.GetResults(&rel)
			paths, _ := slices2.CollectErr(arr.IterDisplayNames(co.SIGDN_FILESYSPATH))

			me.withWaitCursor(func() {
				me.addMp3sToList(paths)
			})
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_EDIT, func() {
		if me.editSelected() == co.ID_OK {
			me.withWaitCursor(func() {
				me.saveSelected()
			})
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_REMOVE, func() {
		me.lstFiles.Items.DeleteSelected()
		me.updateTitlebarCount()
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_RESAVE, func() {
		ret, _ := win.TaskDialogIndirect(win.TASKDIALOGCONFIG{
			HwndParent:  me.wnd.Hwnd(),
			WindowTitle: "Save files",
			Content:     fmt.Sprintf("Do you want to rewrite the tags of %d file(s)?", me.lstFiles.Items.SelectedCount()),
			HMainIcon:   win.TdcIconTdi(co.TDICON_WARNING),
			Flags:       co.TDF_ALLOW_DIALOG_CANCELLATION | co.TDF_POSITION_RELATIVE_TO_WINDOW,
			Buttons: []win.TASKDIALOG_BUTTON{
				{Id: co.ID_OK, Text: "&Save"},
				{Id: co.ID_CANCEL, Text: "&Cancel"},
			},
		})
		if ret == co.ID_OK {
			me.withWaitCursor(func() {
				me.saveSelected()
			})
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_DELPIC, func() {
		me.removePicRg(false)
		me.withWaitCursor(func() {
			me.saveSelected()
		})
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_DELPICRG, func() {
		me.removePicRg(true)
		me.withWaitCursor(func() {
			me.saveSelected()
		})
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FILE_ABOUT, func() {
		hInst, _ := win.GetModuleHandle("")
		exeName, _ := hInst.GetModuleFileName()
		nfo, _ := win.VersionLoad(exeName)

		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)

		caption := fmt.Sprintf("%s %d.%d.%d",
			nfo.ProductName, nfo.Version[0], nfo.Version[1], nfo.Version[2])

		msg := fmt.Sprintf(
			"%s\n\n"+
				"Compiler: %s\n"+
				"GC cycles: %d\n"+
				"Alloc: %s\n"+
				"Next GC: %s\n"+
				"Frees: %d",
			nfo.LegalCopyright,
			runtime.Version(),
			stats.NumGC, win.Str.FmtBytes(stats.HeapAlloc),
			win.Str.FmtBytes(stats.NextGC), stats.Frees)

		win.TaskDialogIndirect(win.TASKDIALOGCONFIG{
			HwndParent:      me.wnd.Hwnd(),
			WindowTitle:     "About",
			MainInstruction: caption,
			Content:         msg,
			HMainIcon:       win.TdcIconTdi(co.TDICON_INFORMATION),
			CommonButtons:   co.TDCBF_OK,
			Flags:           co.TDF_ALLOW_DIALOG_CANCELLATION | co.TDF_POSITION_RELATIVE_TO_WINDOW,
		})
	})

	me.lstFiles.On().NmDblClk(func(_ *win.NMITEMACTIVATE) {
		if me.editSelected() == co.ID_OK {
			me.withWaitCursor(func() {
				me.saveSelected()
			})
		}
	})

	me.lstFiles.On().LvnKeyDown(func(p *win.NMLVKEYDOWN) {
		if p.WVKey == co.VK_DELETE {
			me.lstFiles.Items.DeleteSelected()
			me.updateTitlebarCount()
		} else if p.WVKey == co.VK_RETURN { // Enter key
			if me.editSelected() == co.ID_OK {
				me.withWaitCursor(func() {
					me.saveSelected()
				})
			}
		}
	})

	me.lstFiles.On().LvnItemChanged(func(_ *win.NMLISTVIEW) {
		me.updateTitlebarCount()
	})

	me.lstFiles.Header().On().HdnItemClick(func(p *win.NMHEADER) {
		newCol := me.lstFiles.Cols.Get(int(p.IItem))
		if newCol.Index() != me.sortCol { // changing column
			newCol.SetSortArrow(co.HDF_SORTUP)
			me.sortAsc = true
		} else { // reversing current column
			curArrow := newCol.SortArrow()
			if curArrow == co.HDF_SORTUP {
				newCol.SetSortArrow(co.HDF_SORTDOWN)
				me.sortAsc = false
			} else {
				newCol.SetSortArrow(co.HDF_SORTUP)
				me.sortAsc = true
			}
		}
		me.sortCol = newCol.Index()
		me.sortList()
	})

}
