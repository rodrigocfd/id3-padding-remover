package main

import (
	"fmt"
	"id3fit/id3"
	"id3fit/prompt"
	"runtime"
	"unsafe"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/com/shell"
	"github.com/rodrigocfd/windigo/win/com/shell/shellco"
)

func createAccelTable() ui.AcceleratorTable {
	return ui.NewAcceleratorTable().
		AddChar('o', co.ACCELF_CONTROL, MNU_OPEN).
		AddKey(co.VK_F1, co.ACCELF_NONE, MNU_ABOUT)
}

func createContextMenu() win.HMENU {
	return win.CreatePopupMenu().
		AddItem(MNU_OPEN, "&Open files...\tCtrl+O").
		AddItem(MNU_DELETE, "&Delete from list\tDel").
		AddSeparator().
		AddItem(MNU_REM_PAD, "Remove &padding").
		AddItem(MNU_REM_RG, "Remove Replay&Gain").
		AddItem(MNU_REM_RG_PIC, "Remove ReplayGain and p&ic").
		AddItem(MNU_PREFIX_YEAR, "Prefix album with &year").
		AddSeparator().
		AddItem(MNU_ABOUT, "&About...\tF1")
}

func (me *DlgMain) eventsMenu() {
	me.wnd.On().WmInitMenuPopup(func(p wm.InitMenuPopup) {
		if p.Hmenu() == me.lstFiles.ContextMenu() {
			p.Hmenu().EnableByCmdId(
				me.lstFiles.Items().SelectedCount() > 0, // 1 or more files currently selected
				MNU_DELETE, MNU_PREFIX_YEAR, MNU_REM_PAD, MNU_REM_RG, MNU_REM_RG_PIC)
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_OPEN, func(_ wm.Command) {
		fod := shell.NewIFileOpenDialog(co.CLSCTX_INPROC_SERVER)
		defer fod.Release()

		flags := fod.GetOptions()
		fod.SetOptions(flags | shellco.FOS_FORCEFILESYSTEM |
			shellco.FOS_FILEMUSTEXIST | shellco.FOS_ALLOWMULTISELECT)

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

			t0 := win.QueryPerformanceCounter()
			me.addFilesToList(mp3s, func() {
				me.tellElapsedTime(t0, len(mp3s))
			})
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_DELETE, func(_ wm.Command) {
		me.lstFiles.SetRedraw(false)
		me.lstFiles.Items().DeleteSelected() // will fire multiple LVM_DELETEITEM
		me.lstFiles.SetRedraw(true)
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_PAD, func(_ wm.Command) {
		t0 := win.QueryPerformanceCounter()
		me.reSaveTagsOfSelectedFiles(func() { // simply saving will remove the padding
			me.tellElapsedTime(t0, me.lstFiles.Items().SelectedCount())
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG, func(_ wm.Command) {
		t0 := win.QueryPerformanceCounter()
		selItems := me.lstFiles.Items().Selected()

		for _, selItem := range selItems {
			tag := me.cachedTags[selItem.Text(0)]
			tag.DeleteFrames(func(fr id3.Frame) bool {
				if frMulti, ok := fr.(*id3.FrameMultiText); ok {
					return frMulti.IsReplayGain()
				}
				return false
			})
		}

		me.reSaveTagsOfSelectedFiles(func() {
			me.tellElapsedTime(t0, len(selItems))
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG_PIC, func(_ wm.Command) {
		t0 := win.QueryPerformanceCounter()
		selItems := me.lstFiles.Items().Selected()

		for _, selItem := range selItems {
			tag := me.cachedTags[selItem.Text(0)]
			tag.DeleteFrames(func(frDyn id3.Frame) bool {
				if frMulti, ok := frDyn.(*id3.FrameMultiText); ok {
					if frMulti.IsReplayGain() {
						return true
					}
				} else if frBin, ok := frDyn.(*id3.FrameBinary); ok {
					if frBin.Name4() == "APIC" {
						return true
					}
				}
				return false
			})
		}

		me.reSaveTagsOfSelectedFiles(func() {
			me.tellElapsedTime(t0, len(selItems))
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_PREFIX_YEAR, func(_ wm.Command) {
		t0 := win.QueryPerformanceCounter()
		selItems := me.lstFiles.Items().Selected()

		for _, selItem := range selItems {
			tag := me.cachedTags[selItem.Text(0)]
			frAlbDyn, hasAlb := tag.FrameByName("TALB")
			frYerDyn, hasYer := tag.FrameByName("TYER")

			if !hasAlb {
				prompt.Error(me.wnd, "Missing frame", "", "Album frame not found.")
			} else if !hasYer {
				prompt.Error(me.wnd, "Missing frame", "", "Year frame not found.")
			}

			frAlb, _ := frAlbDyn.(*id3.FrameText)
			frYer, _ := frYerDyn.(*id3.FrameText)
			*frAlb.Text() = fmt.Sprintf("%s %s", *frYer.Text(), *frAlb.Text())
		}

		me.reSaveTagsOfSelectedFiles(func() {
			me.tellElapsedTime(t0, len(selItems))
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_ABOUT, func(_ wm.Command) {
		resourceSl := win.GetFileVersionInfo(win.HINSTANCE(0).GetModuleFileName())
		versionSl, _ := win.VerQueryValue(resourceSl, "\\")
		vsffi := (*win.VS_FIXEDFILEINFO)(unsafe.Pointer(&versionSl[0]))
		v := vsffi.ProductVersion()

		memStats := runtime.MemStats{}
		runtime.ReadMemStats(&memStats)

		prompt.Info(me.wnd, "About",
			fmt.Sprintf("ID3 Fit %d.%d.%d", v[0], v[1], v[2]),
			fmt.Sprintf("Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Windigo library.\n\n"+
				"Alloc mem: %s\n"+
				"Alloc sys: %s\n"+
				"Alloc idle: %s\n"+
				"GC cycles: %d\n"+
				"Next GC: %s\n"+
				"Alloc objs: %d",
				win.Str.FmtBytes(memStats.HeapAlloc),
				win.Str.FmtBytes(memStats.HeapSys),
				win.Str.FmtBytes(memStats.HeapIdle),
				memStats.NumGC,
				win.Str.FmtBytes(memStats.NextGC),
				memStats.HeapObjects,
			))
	})
}
