package main

import (
	"fmt"
	"id3fit/id3"
	"sort"

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
	hMenu := win.CreatePopupMenu()

	hMenu.AddItem(MNU_OPEN, "&Open files...\tCtrl+O")
	hMenu.AddItem(MNU_DELETE, "&Delete from list\tDel")
	hMenu.AddSeparator()
	hMenu.AddItem(MNU_REM_PAD, "Remove &padding")
	hMenu.AddItem(MNU_REM_RG, "Remove Replay&Gain")
	hMenu.AddItem(MNU_REM_RG_PIC, "Remove ReplayGain and p&ic")
	hMenu.AddItem(MNU_PREFIX_YEAR, "Prefix album with &year")
	hMenu.AddSeparator()
	hMenu.AddItem(MNU_ABOUT, "&About...\tF1")

	return hMenu
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
		fod := shell.CoCreateIFileOpenDialog(co.CLSCTX_INPROC_SERVER)
		defer fod.Release()

		flags := fod.GetOptions()
		fod.SetOptions(flags | shellco.FOS_FORCEFILESYSTEM |
			shellco.FOS_FILEMUSTEXIST | shellco.FOS_ALLOWMULTISELECT)

		fod.SetFileTypes([]shell.FilterSpec{
			{Name: "MP3 audio files", Spec: "*.mp3"},
			{Name: "All files", Spec: "*.*"},
		})
		fod.SetFileTypeIndex(0)

		if fod.Show(me.wnd.Hwnd()) {
			shia := fod.GetResults()
			defer shia.Release()

			mp3s := shia.GetDisplayNames(shellco.SIGDN_FILESYSPATH)
			sort.Strings(mp3s)
			me.addFilesToList(mp3s)
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_DELETE, func(_ wm.Command) {
		me.lstFiles.SetRedraw(false)
		me.lstFiles.Items().DeleteSelected() // will fire multiple LVM_DELETEITEM
		me.lstFiles.SetRedraw(true)
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_PAD, func(_ wm.Command) {
		me.measureFileProcess(func() {
			me.reSaveTagsOfSelectedFiles() // simply saving will remove the padding
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG, func(_ wm.Command) {
		me.measureFileProcess(func() {
			for _, selItem := range me.lstFiles.Items().Selected() {
				tag := me.cachedTags[selItem.Text(0)]
				tag.DeleteFrames(func(fr id3.Frame) bool {
					if frMulti, ok := fr.(*id3.FrameMultiText); ok {
						return frMulti.IsReplayGain()
					}
					return false
				})
			}

			me.reSaveTagsOfSelectedFiles()
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG_PIC, func(_ wm.Command) {
		me.measureFileProcess(func() {
			for _, selItem := range me.lstFiles.Items().Selected() {
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

			me.reSaveTagsOfSelectedFiles()
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_PREFIX_YEAR, func(_ wm.Command) {
		me.measureFileProcess(func() {
			for _, selItem := range me.lstFiles.Items().Selected() {
				tag := me.cachedTags[selItem.Text(0)]
				frAlbDyn, hasAlb := tag.FrameByName("TALB")
				frYerDyn, hasYer := tag.FrameByName("TYER")

				if !hasAlb {
					me.wnd.Hwnd().TaskDialog(0, APP_TITLE, "Missing frame",
						"Album frame not found.", co.TDCBF_OK, co.TD_ICON_ERROR)
				} else if !hasYer {
					me.wnd.Hwnd().TaskDialog(0, APP_TITLE, "Missing frame",
						"Year frame not found.", co.TDCBF_OK, co.TD_ICON_ERROR)
				}

				frAlb, _ := frAlbDyn.(*id3.FrameText)
				frYer, _ := frYerDyn.(*id3.FrameText)
				*frAlb.Text() = fmt.Sprintf("%s %s", *frYer.Text(), *frAlb.Text())
			}

			me.reSaveTagsOfSelectedFiles()
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_ABOUT, func(_ wm.Command) {
		me.wnd.Hwnd().TaskDialog(0, APP_TITLE, "About",
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Windigo library.",
			co.TDCBF_OK, co.TD_ICON_INFORMATION)
	})
}
