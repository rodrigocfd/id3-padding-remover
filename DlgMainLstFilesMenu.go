package main

import (
	"fmt"
	"id3fit/id3"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/com/shell"
)

const (
	MNU_OPEN int = iota + 1001
	MNU_DELETE
	MNU_REM_PAD
	MNU_REM_RG
	MNU_REM_RG_PIC
	MNU_PREFIX_YEAR
	MNU_ABOUT
)

func createAccelTable() ui.AcceleratorTable {
	return ui.NewAcceleratorTable().
		AddChar('o', co.ACCELF_CONTROL, MNU_OPEN).
		AddKey(co.VK_DELETE, co.ACCELF_NONE, MNU_DELETE).
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

func (me *DlgMain) eventsLstFilesMenu() {
	me.wnd.On().WmInitMenuPopup(func(p wm.InitMenuPopup) {
		if p.Hmenu() == me.lstFiles.ContextMenu() {
			p.Hmenu().EnableByCmdId(
				me.lstFiles.Items().SelectedCount() > 0, // 1 or more files currently selected
				MNU_DELETE, MNU_PREFIX_YEAR, MNU_REM_PAD, MNU_REM_RG, MNU_REM_RG_PIC)
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_OPEN, func(_ wm.Command) {
		mp3s, ok := ui.Prompt.OpenMultipleFiles(me.wnd,
			[]shell.FilterSpec{
				{Name: "MP3 audio files", Spec: "*.mp3"},
				{Name: "All files", Spec: "*.*"},
			})

		if ok {
			me.addFilesToList(mp3s)
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_DELETE, func(_ wm.Command) {
		me.lstFiles.SetRedraw(false)
		me.lstFiles.Items().Delete(me.lstFiles.Items().Selected()...) // will fire LVM_DELETEITEM
		me.lstFiles.SetRedraw(true)
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_PAD, func(_ wm.Command) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {}) // simply saving will remove the padding
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG, func(_ wm.Command) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			tag.DeleteFrames(func(fr id3.Frame) bool {
				if frMulti, ok := fr.(*id3.FrameMultiText); ok {
					return frMulti.IsReplayGain()
				}
				return false
			})
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG_PIC, func(_ wm.Command) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
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
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_PREFIX_YEAR, func(_ wm.Command) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			frAlbDyn := tag.FrameByName("TALB")
			frYerDyn := tag.FrameByName("TYER")

			if frAlbDyn == nil {
				ui.Prompt.MessageBox(me.wnd, "Album frame not found.",
					"Missing frame", co.MB_ICONERROR)
			} else if frYerDyn == nil {
				ui.Prompt.MessageBox(me.wnd, "Year frame not found.",
					"Missing frame", co.MB_ICONERROR)
			}

			if frAlb, ok := frAlbDyn.(*id3.FrameText); ok {
				if frYer, ok := frYerDyn.(*id3.FrameText); ok {
					frAlb.SetText(fmt.Sprintf("%s %s", frYer.Text(), frAlb.Text()))
				}
			}
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_ABOUT, func(_ wm.Command) {
		ui.Prompt.MessageBox(me.wnd,
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Windigo library.",
			"About", co.MB_ICONINFORMATION)
	})
}
