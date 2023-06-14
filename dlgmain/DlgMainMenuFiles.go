package dlgmain

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/timecount"
	"runtime"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/com/com"
	"github.com/rodrigocfd/windigo/win/com/com/comco"
	"github.com/rodrigocfd/windigo/win/com/shell"
	"github.com/rodrigocfd/windigo/win/com/shell/shellco"
)

func (me *DlgMain) prepareContextMenuFiles(p wm.InitMenuPopup) {
	atLeastOneSel := me.lstMp3s.Items().SelectedCount() > 0
	cmdIds := []int{MNU_MP3_DELETE, // these menu entries are enabled if at least 1 file is selected
		MNU_MP3_REM_PAD, MNU_MP3_REM_RG, MNU_MP3_REM_RG_PIC, MNU_MP3_DEL_TAG,
		MNU_MP3_COPY_TO_FOLDER, MNU_MP3_RENAME, MNU_MP3_RENAME_PREFIX}
	for _, cmdId := range cmdIds {
		p.Hmenu().EnableMenuItem(win.MenuItemCmd(cmdId), atLeastOneSel)
	}
}

func (me *DlgMain) eventsMenuFiles() {

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_OPEN, func(_ wm.Command) {
		fod := shell.NewIFileOpenDialog(
			com.CoCreateInstance(
				shellco.CLSID_FileOpenDialog, nil,
				comco.CLSCTX_INPROC_SERVER,
				shellco.IID_IFileOpenDialog),
		)
		defer fod.Release()

		fod.SetOptions(fod.GetOptions() |
			shellco.FOS_FORCEFILESYSTEM |
			shellco.FOS_FILEMUSTEXIST |
			shellco.FOS_ALLOWMULTISELECT)

		fod.SetFileTypes([]shell.FilterSpec{
			{Name: "MP3 audio files", Spec: "*.mp3"},
			{Name: "All files", Spec: "*.*"},
		})
		fod.SetFileTypeIndex(1)

		if fod.Show(me.wnd.Hwnd()) {
			mp3s := fod.ListResultDisplayNames(shellco.SIGDN_FILESYSPATH)
			if me.modalTagOp(TAG_OP_LOAD, mp3s...) {
				me.addMp3sToList(mp3s)
			}
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_DELETE, func(_ wm.Command) {
		me.lstMp3s.SetRedraw(false)
		me.lstMp3s.Items().DeleteSelected() // will fire multiple LVM_DELETEITEM
		me.lstMp3s.SetRedraw(true)
		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_REM_PAD, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3s := me.lstMp3s.Columns().Get(0).SelectedTexts()

		// Simply saving will remove the padding.
		if me.modalTagOp(TAG_OP_SAVE_AND_RELOAD, selMp3s...) {
			me.addMp3sToList(selMp3s)

			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("Padding removed from %d file(s) in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_REM_RG, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3s := me.lstMp3s.Columns().Get(0).SelectedTexts()

		for _, selMp3 := range selMp3s {
			tag := me.cachedTags[selMp3]
			tag.DeleteFrames(func(_ int, frame *id3v2.Frame) (willDelete bool) {
				return frame.IsReplayGain()
			})
		}

		if me.modalTagOp(TAG_OP_SAVE_AND_RELOAD, selMp3s...) {
			me.addMp3sToList(selMp3s)
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("ReplayGain removed from %d file(s) in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_REM_RG_PIC, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3s := me.lstMp3s.Columns().Get(0).SelectedTexts()

		for _, selMp3 := range selMp3s {
			tag := me.cachedTags[selMp3]
			tag.DeleteFrames(func(_ int, frame *id3v2.Frame) (willDelete bool) {
				return frame.IsReplayGain() || frame.Name4() == "APIC"
			})
		}

		if me.modalTagOp(TAG_OP_SAVE_AND_RELOAD, selMp3s...) {
			me.addMp3sToList(selMp3s)
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("ReplayGain and album art removed from %d file(s) in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_DEL_TAG, func(_ wm.Command) {
		selMp3s := me.lstMp3s.Columns().Get(0).SelectedTexts()

		proceed := ui.TaskDlg.OkCancelEx(me.wnd, "Delete tag", win.StrOptNone(),
			fmt.Sprintf("Permanently delete the tag from %d file(s)?", len(selMp3s)),
			win.StrOptSome("Delete"), win.StrOptNone())
		if !proceed {
			return
		}

		t0 := timecount.New()

		for _, selMp3 := range selMp3s {
			tag := me.cachedTags[selMp3]
			tag.DeleteFrames(func(_ int, _ *id3v2.Frame) (willDelete bool) {
				return true
			})
		}

		if me.modalTagOp(TAG_OP_SAVE_AND_RELOAD, selMp3s...) {
			me.addMp3sToList(selMp3s)
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("Tag deleted from %d file(s) in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_COPY_TO_FOLDER, func(_ wm.Command) {
		fod := shell.NewIFileOpenDialog(
			com.CoCreateInstance(
				shellco.CLSID_FileOpenDialog, nil,
				comco.CLSCTX_INPROC_SERVER,
				shellco.IID_IFileOpenDialog),
		)
		defer fod.Release()

		fod.SetOptions(fod.GetOptions() | shellco.FOS_PICKFOLDERS)
		if !fod.Show(me.wnd.Hwnd()) {
			return
		}

		newFolder := fod.GetResultDisplayName(shellco.SIGDN_FILESYSPATH)
		selMp3s := me.lstMp3s.Columns().Get(0).SelectedTexts()
		newCopiedFiles := make([]string, 0, len(selMp3s))
		t0 := timecount.New()

		for _, selMp3 := range selMp3s {
			newPath := fmt.Sprintf("%s\\%s",
				newFolder, win.Path.GetFileName(selMp3))
			if win.Path.Exists(newPath) {
				ui.TaskDlg.Error(me.wnd, "Existing file", win.StrOptNone(),
					fmt.Sprintf("File already exists:\n%s", newPath))
				return
			}
			if err := win.CopyFile(selMp3, newPath, false); err != nil {
				ui.TaskDlg.Error(me.wnd, "Copy error", win.StrOptNone(), err.Error())
				return
			}
			newCopiedFiles = append(newCopiedFiles, newPath)
		}

		if len(newCopiedFiles) == 0 {
			ui.TaskDlg.Info(me.wnd, "No copies", win.StrOptNone(), "No files have been copied.")
			return
		}

		me.lstMp3s.SetRedraw(false)
		for _, selMp3 := range selMp3s { // delete all items of the copied files
			item, _ := me.lstMp3s.Items().Find(selMp3)
			item.Delete() // will fire LVM_DELETEITEM
		}
		me.lstMp3s.SetRedraw(true)

		if me.modalTagOp(TAG_OP_LOAD, newCopiedFiles...) {
			me.addMp3sToList(newCopiedFiles) // load the files that have been copied to the new folder
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d file(s) reloaded in %.2f ms.",
					len(newCopiedFiles), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_RENAME, func(_ wm.Command) {
		t0 := timecount.New()
		if count, err := me.renameSelectedFiles(false); err != nil {
			ui.TaskDlg.Error(me.wnd, "Renaming error", win.StrOptNone(), "Error: "+err.Error())
		} else {
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d file(s) renamed in %.2f ms.",
					count, t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_RENAME_PREFIX, func(_ wm.Command) {
		t0 := timecount.New()
		if count, err := me.renameSelectedFiles(true); err != nil {
			ui.TaskDlg.Error(me.wnd, "Renaming error", win.StrOptNone(), "Error: "+err.Error())
		} else {
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d file(s) renamed in %.2f ms.",
					count, t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_MP3_ABOUT, func(_ wm.Command) {
		resNfo, _ := win.ResourceInfoLoad(win.HINSTANCE(0).GetModuleFileName())
		vMaj, vMin, vPat, _ := resNfo.ProductVersion()

		blocks := resNfo.Blocks()
		productName, _ := blocks[0].ProductName()

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		ui.TaskDlg.Info(me.wnd, "About",
			win.StrOptSome(fmt.Sprintf("%s %d.%d.%d", productName, vMaj, vMin, vPat)),
			fmt.Sprintf("Rodrigo César de Freitas Dias (C) 2021\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go %s with Windigo library.\n\n"+
				"Objects mem: %s\n"+
				"Reserved sys: %s\n"+
				"Idle spans: %s\n"+
				"GC cycles: %d\n"+
				"Next GC: %s",
				strings.TrimLeft(runtime.Version(), "go"),
				win.Str.FmtBytes(memStats.HeapAlloc),
				win.Str.FmtBytes(memStats.HeapSys),
				win.Str.FmtBytes(memStats.HeapIdle),
				memStats.NumGC,
				win.Str.FmtBytes(memStats.NextGC),
			))

		me.updateMemoryStatus()
	})
}
