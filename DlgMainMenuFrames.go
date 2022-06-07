package main

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
)

func (me *DlgMain) initMenuPopupFrames(p wm.InitMenuPopup) {
	atLeastOneSel := me.lstFrames.Items().SelectedCount() > 0
	firstIsSel := me.lstFrames.Items().Get(0).IsSelected()

	isApicSelected := false
	for _, selItem := range me.lstFrames.Items().SelectedItems() {
		if selItem.Text(0) == "APIC" {
			isApicSelected = true
			break
		}
	}

	p.Hmenu().EnableMenuItem(win.MenuItemCmd(MNU_FRAMES_MOVEUP), atLeastOneSel && !firstIsSel && !isApicSelected)
	p.Hmenu().EnableMenuItem(win.MenuItemCmd(MNU_FRAMES_DEL), atLeastOneSel)
}

func (me *DlgMain) eventsMenuFrames() {

	me.wnd.On().WmCommandAccelMenu(MNU_FRAMES_MOVEUP, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3 := me.lstMp3s.Items().SelectedItems()[0].Text(0) // single selected MP3 file
		tag := me.cachedTags[selMp3]
		idxsToMove := me.lstFrames.Items().SelectedIndexes()

		// Assumes validation has been made on WM_INITMENUPOPUP,
		// so no invalid frames are selected.

		for _, idxToMove := range idxsToMove { // swap each selected frame within the Frames slice
			tag.SwapFrames(idxToMove, idxToMove-1)
		}

		if me.modalTagOp([]string{selMp3}, TAG_OP_SAVE_AND_RELOAD) {
			me.displayFramesOfSelectedFiles()
			for _, idxToMove := range idxsToMove { // restore the selected items
				me.lstFrames.Items().Get(idxToMove - 1).Select(true)
			}
			me.lstFrames.Focus()

			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d frame(s) moved up in %.2f ms.",
					len(idxsToMove), t0.ElapsedMs()))
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FRAMES_DEL, func(_ wm.Command) {
		selMp3 := me.lstMp3s.Columns().Get(0).SelectedTexts()[0] // single selected MP3 file
		tag := me.cachedTags[selMp3]
		idxsToDelete := me.lstFrames.Items().SelectedIndexes()

		proceed := ui.TaskDlg.OkCancelEx(me.wnd, "Delete frames", win.StrOptNone(),
			fmt.Sprintf("Are you sure you want to delete %d frame(s)?", len(idxsToDelete)),
			win.StrOptSome("Delete"), win.StrOptNone())
		if !proceed {
			return
		}

		t0 := timecount.New()
		tag.DeleteFrames(func(i int, _ *id3v2.Frame) (willDelete bool) {
			for _, idxToDelete := range idxsToDelete {
				if i == idxToDelete {
					return true
				}
			}
			return false
		})

		if me.modalTagOp([]string{selMp3}, TAG_OP_SAVE_AND_RELOAD) {
			me.displayFramesOfSelectedFiles()
			ui.TaskDlg.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d frame(s) deleted from tag in %.2f ms.",
					len(idxsToDelete), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})
}
