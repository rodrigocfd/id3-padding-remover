package main

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/prompt"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
)

func (me *DlgMain) initMenuPopupFrames(p wm.InitMenuPopup) {
	atLeastOneSel := me.lstFrames.Items().SelectedCount() > 0
	cmdIds := []int{MNU_FRAMES_MOVEUP, MNU_FRAMES_REM}
	for _, cmdId := range cmdIds {
		p.Hmenu().EnableMenuItem(win.MenuItemCmd(cmdId), atLeastOneSel)
	}
}

func (me *DlgMain) eventsMenuFrames() {

	me.wnd.On().WmCommandAccelMenu(MNU_FRAMES_MOVEUP, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3 := me.lstMp3s.Items().SelectedItems()[0].Text(0) // single selected MP3 file
		tag := me.cachedTags[selMp3]
		idxsToMove := me.lstFrames.Items().SelectedIndexes()

		if idxsToMove[0] == 0 {
			prompt.Error(me.wnd, "Bad move", win.StrOptNone(), "First item cannot be moved up.")
			return
		}

		for _, idxToMove := range idxsToMove { // swap each selected frame within the Frames slice
			tmp := tag.Frames()[idxToMove-1]
			tag.Frames()[idxToMove-1] = tag.Frames()[idxToMove]
			tag.Frames()[idxToMove] = tmp
		}

		if me.modalTagOp([]string{selMp3}, TAG_OP_SAVE_AND_RELOAD) {
			me.displayFramesOfSelectedFiles()
			for _, idxToMove := range idxsToMove { // restore the selected items
				me.lstFrames.Items().Get(idxToMove - 1).Select(true)
			}
			me.lstFrames.Focus()

			prompt.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d frame(s) moved up in %.2f ms.",
					len(idxsToMove), t0.ElapsedMs()))
		}

		me.updateMemoryStatus()
	})

	me.wnd.On().WmCommandAccelMenu(MNU_FRAMES_REM, func(_ wm.Command) {
		selMp3 := me.lstMp3s.Columns().SelectedTexts(0)[0] // single selected MP3 file
		tag := me.cachedTags[selMp3]
		idxsToDelete := me.lstFrames.Items().SelectedIndexes()

		if !prompt.OkCancel(me.wnd, "Deleting frames", win.StrOptNone(),
			fmt.Sprintf("Are you sure you want to delete %d frame(s)?", len(idxsToDelete))) {
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
			prompt.Info(me.wnd, "Process finished", win.StrOptSome("Success"),
				fmt.Sprintf("%d frame(s) deleted from tag in %.2f ms.",
					len(idxsToDelete), t0.ElapsedMs()))
			me.lstMp3s.Focus()
		}

		me.updateMemoryStatus()
	})
}
