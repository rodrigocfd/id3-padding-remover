package main

import (
	"fmt"
	"id3-fit/id3"
	"windigo/co"
	"windigo/com/shell"
	"windigo/ui"
)

func (me *DlgMain) buildLstFilesMenuAndAccel() {
	me.lstFilesMenu.
		AppendItem(MNU_OPEN, "&Open files...\tCtrl+O").
		AppendItem(MNU_DELETE, "&Delete from list\tDel").
		AppendSeparator().
		AppendItem(MNU_REM_PAD, "Remove &padding").
		AppendItem(MNU_REM_RG, "Remove Replay&Gain").
		AppendItem(MNU_REM_RG_PIC, "Remove ReplayGain and p&ic").
		AppendItem(MNU_PREFIX_YEAR, "Prefix album with &year").
		AppendSeparator().
		AppendItem(MNU_ABOUT, "&About...\tF1")

	me.wnd.AccelTable().
		AddChar('o', co.ACCELF_CONTROL, MNU_OPEN).
		AddKey(co.VK_DELETE, co.ACCELF_NONE, MNU_DELETE).
		AddKey(co.VK_F1, co.ACCELF_NONE, MNU_ABOUT)
}

func (me *DlgMain) eventsLstFilesMenu() {
	me.wnd.On().WmInitMenuPopup(func(p ui.WmInitMenuPopup) {
		if p.Hmenu() == me.lstFilesMenu.Hmenu() {
			me.lstFilesMenu.EnableItemsByCmdId(
				me.lstFiles.Items().SelectedCount() > 0, // 1 or more files currently selected
				MNU_DELETE, MNU_PREFIX_YEAR, MNU_REM_PAD, MNU_REM_RG, MNU_REM_RG_PIC,
			)
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_OPEN, func(_ ui.WmCommand) {
		mp3s, ok := ui.SysDlg.OpenMultipleFiles(me.wnd,
			[]shell.FilterSpec{
				{Name: "MP3 audio files", Spec: "*.mp3"},
				{Name: "All files", Spec: "*.*"},
			})

		if ok {
			me.addFilesToList(mp3s)
		}
	})

	me.wnd.On().WmCommandAccelMenu(MNU_DELETE, func(_ ui.WmCommand) {
		me.lstFiles.SetRedraw(false).
			Items().DeleteSelected(). // will fire LVM_DELETEITEM
			SetRedraw(true)
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_PAD, func(_ ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {}) // simply saving will remove the padding
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG, func(_ ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			tag.DeleteFrames(func(fr id3.Frame) bool {
				if frMulti, ok := fr.(*id3.FrameMultiText); ok {
					return frMulti.IsReplayGain()
				}
				return false
			})
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_REM_RG_PIC, func(_ ui.WmCommand) {
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

	me.wnd.On().WmCommandAccelMenu(MNU_PREFIX_YEAR, func(_ ui.WmCommand) {
		me.reSaveTagsOfSelectedFiles(func(tag *id3.Tag) {
			frAlbDyn := tag.FrameByName("TALB")
			frYerDyn := tag.FrameByName("TYER")

			if frAlbDyn == nil {
				ui.SysDlg.MsgBox(me.wnd, "Album frame not found.", "Missing frame", co.MB_ICONERROR)
			} else if frYerDyn == nil {
				ui.SysDlg.MsgBox(me.wnd, "Year frame not found.", "Missing frame", co.MB_ICONERROR)
			}

			if frAlb, ok := frAlbDyn.(*id3.FrameText); ok {
				if frYer, ok := frYerDyn.(*id3.FrameText); ok {
					frAlb.SetText(fmt.Sprintf("%s %s", frYer.Text(), frAlb.Text()))
				}
			}
		})
	})

	me.wnd.On().WmCommandAccelMenu(MNU_ABOUT, func(_ ui.WmCommand) {
		ui.SysDlg.MsgBox(me.wnd,
			"ID3 Fit 2.0.0\n"+
				"Rodrigo César de Freitas Dias\n"+
				"rcesar@gmail.com\n\n"+
				"This application was written in Go with Windigo library.",
			"About", co.MB_ICONINFORMATION)
	})
}
