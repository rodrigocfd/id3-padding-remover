package main

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/prompt"
	"id3fit/timecount"
	"runtime"

	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/com/shell"
	"github.com/rodrigocfd/windigo/win/com/shell/shellco"
)

func (me *DlgMain) eventsMenu() {
	me.wnd.On().WmInitMenuPopup(func(p wm.InitMenuPopup) {
		if p.Hmenu() == me.lstMp3s.ContextMenu() {
			cmdIds := []int{MNU_DELETE,
				MNU_REM_PAD, MNU_REM_RG, MNU_REM_RG_PIC, MNU_DEL_TAG,
				MNU_COPY, MNU_RENAME, MNU_RENAME_PREFIX}
			for _, cmdId := range cmdIds {
				p.Hmenu().EnableMenuItem(win.MenuItemCmd(cmdId),
					me.lstMp3s.Items().SelectedCount() > 0) // 1 or more files currently selected
			}
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_OPEN, func(_ wm.Command) {
		fod := shell.NewIFileOpenDialog(
			win.CoCreateInstance(
				shellco.CLSID_FileOpenDialog, nil,
				co.CLSCTX_INPROC_SERVER,
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

		// shiDir, _ := shell.NewShellItem(win.GetCurrentDirectory())
		// defer shiDir.Release()
		// fod.SetFolder(&shiDir)

		if fod.Show(me.wnd.Hwnd()) {
			mp3s := fod.GetResultsDisplayNames(shellco.SIGDN_FILESYSPATH)
			win.Path.Sort(mp3s)

			// t0 := timecount.New()
			me.addFilesToList(mp3s, func() {
				// prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
				// 	fmt.Sprintf("%d file tag(s) parsed in %.2f ms.",
				// 		len(mp3s), t0.ElapsedMs()))
			})
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_DELETE, func(_ wm.Command) {
		me.lstMp3s.SetRedraw(false)
		me.lstMp3s.Items().DeleteSelected() // will fire multiple LVM_DELETEITEM
		me.lstMp3s.SetRedraw(true)
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_PAD, func(_ wm.Command) {
		t0 := timecount.New()
		me.reSaveTagsOfSelectedFiles(func() { // simply saving will remove the padding
			prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
				fmt.Sprintf("Padding removed from %d file(s) in %.2f ms.",
					me.lstMp3s.Items().SelectedCount(), t0.ElapsedMs()))
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3s := me.lstMp3s.Columns().SelectedTexts(0)

		for _, selMp3 := range selMp3s {
			tag := me.cachedTags[selMp3]
			tag.DeleteFrames(func(fr id3v2.Frame) (willDelete bool) {
				if frMulti, ok := fr.(*id3v2.FrameMultiText); ok {
					return frMulti.IsReplayGain()
				}
				return false
			})
		}

		me.reSaveTagsOfSelectedFiles(func() {
			prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
				fmt.Sprintf("ReplayGain removed from %d file(s) in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG_PIC, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3s := me.lstMp3s.Columns().SelectedTexts(0)

		for _, selMp3 := range selMp3s {
			tag := me.cachedTags[selMp3]
			tag.DeleteFrames(func(frDyn id3v2.Frame) (willDelete bool) {
				if frMulti, ok := frDyn.(*id3v2.FrameMultiText); ok {
					if frMulti.IsReplayGain() {
						return true
					}
				} else if frBin, ok := frDyn.(*id3v2.FrameBinary); ok {
					if frBin.Name4() == "APIC" {
						return true
					}
				}
				return false
			})
		}

		me.reSaveTagsOfSelectedFiles(func() {
			prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
				fmt.Sprintf("ReplayGain and album art removed from %d file(s) in %.2f ms.",
					len(selMp3s), t0.ElapsedMs()))
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_DEL_TAG, func(_ wm.Command) {
		selMp3s := me.lstMp3s.Columns().SelectedTexts(0)
		proceed := prompt.OkCancel(me.wnd, "Delete tag", nil,
			fmt.Sprintf("Completely remove the tag from %d file(s)?", len(selMp3s)))

		if proceed {
			t0 := timecount.New()

			for _, selMp3 := range selMp3s {
				tag := me.cachedTags[selMp3]
				tag.DeleteFrames(func(_ id3v2.Frame) (willDelete bool) {
					return true
				})
			}

			me.reSaveTagsOfSelectedFiles(func() {
				prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
					fmt.Sprintf("Tag deleted from %d file(s) in %.2f ms.",
						len(selMp3s), t0.ElapsedMs()))
			})
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_COPY, func(_ wm.Command) {
		fod := shell.NewIFileOpenDialog(
			win.CoCreateInstance(
				shellco.CLSID_FileOpenDialog, nil,
				co.CLSCTX_INPROC_SERVER,
				shellco.IID_IFileOpenDialog),
		)
		defer fod.Release()

		fod.SetOptions(fod.GetOptions() | shellco.FOS_PICKFOLDERS)

		if fod.Show(me.wnd.Hwnd()) {
			newFolder := fod.GetResultDisplayName(shellco.SIGDN_FILESYSPATH)
			var newCopiedFiles []string
			t0 := timecount.New()

			for _, selMp3 := range me.lstMp3s.Columns().SelectedTexts(0) {
				newPath := fmt.Sprintf("%s\\%s",
					newFolder, win.Path.GetFileName(selMp3))
				if win.Path.Exists(newPath) {
					prompt.Error(me.wnd, "File already exists", nil,
						fmt.Sprintf("File already exists:\n%s", newPath))
					continue
				}
				if err := win.CopyFile(selMp3, newPath, false); err != nil {
					prompt.Error(me.wnd, "File copy error", nil, err.Error())
					continue
				}
				newCopiedFiles = append(newCopiedFiles, newPath)
			}

			me.addFilesToList(newCopiedFiles, func() {
				if len(newCopiedFiles) > 0 {
					prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
						fmt.Sprintf("%d file(s) copied and parsed back in %.2f ms.",
							len(newCopiedFiles), t0.ElapsedMs()))
				}
			})
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_RENAME, func(_ wm.Command) {
		t0 := timecount.New()
		if count, err := me.renameSelectedFiles(false); err != nil {
			prompt.Error(me.wnd, "Renaming error", nil, "Error: "+err.Error())
		} else {
			prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
				fmt.Sprintf("%d file(s) renamed in %.2f ms.",
					count, t0.ElapsedMs()))
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_RENAME_PREFIX, func(_ wm.Command) {
		t0 := timecount.New()
		if count, err := me.renameSelectedFiles(true); err != nil {
			prompt.Error(me.wnd, "Renaming error", nil, "Error: "+err.Error())
		} else {
			prompt.Info(me.wnd, "Process finished", win.StrVal("Success"),
				fmt.Sprintf("%d file(s) renamed in %.2f ms.",
					count, t0.ElapsedMs()))
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_ABOUT, func(_ wm.Command) {
		resNfo, _ := win.ResourceInfoLoad(win.HINSTANCE(0).GetModuleFileName())
		vsf, _ := resNfo.FixedFileInfo()
		vMaj, vMin, vPat, _ := vsf.ProductVersion()

		block0 := resNfo.Blocks()[0]
		productName, _ := resNfo.ProductName(block0.LangId, block0.CodePage)

		memStats := runtime.MemStats{}
		runtime.ReadMemStats(&memStats)

		prompt.Info(me.wnd, "About",
			win.StrVal(fmt.Sprintf("%s %d.%d.%d", productName, vMaj, vMin, vPat)),
			fmt.Sprintf("Rodrigo César de Freitas Dias (C) 2021\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Windigo library.\n\n"+
				"Objects mem: %s\n"+
				"Reserved sys: %s\n"+
				"Idle spans: %s\n"+
				"GC cycles: %d\n"+
				"Next GC: %s",
				win.Str.FmtBytes(memStats.HeapAlloc),
				win.Str.FmtBytes(memStats.HeapSys),
				win.Str.FmtBytes(memStats.HeapIdle),
				memStats.NumGC,
				win.Str.FmtBytes(memStats.NextGC),
			))
	})
}
