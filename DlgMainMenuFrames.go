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
	selFrameNames4 := make([]string, 0, me.lstFrames.Items().SelectedCount())
	for _, name4 := range me.lstFrames.Columns().SelectedTexts(0) {
		if name4 != "" {
			selFrameNames4 = append(selFrameNames4, name4) // only non-empty names
		}
	}
	p.Hmenu().EnableMenuItem(
		win.MenuItemCmd(MNU_FRAMES_REM), len(selFrameNames4) > 0)
}

func (me *DlgMain) eventsMenuFrames() {
	me.wnd.On().WmCommandAccelMenu(MNU_FRAMES_REM, func(_ wm.Command) {
		t0 := timecount.New()
		selMp3 := me.lstMp3s.Columns().SelectedTexts(0)[0] // single selected MP3 file
		tag := me.cachedTags[selMp3]
		idxsToDelete := make([]int, 0, me.lstFrames.Items().SelectedCount())

		selFrameItems := me.lstFrames.Items().Selected()
		for _, selFrameItem := range selFrameItems { // scan all lines of frames listview
			idxFrame := selFrameItem.Index() // index of frame within frames slice
			name4 := selFrameItem.Text(0)
			selFrame := tag.Frames()[idxFrame]

			if selFrame.Name4() != name4 { // additional security check
				prompt.Error(me.wnd, "This is bad", win.StrOptSome("Mismatched frames"),
					fmt.Sprintf("Mismatched frame names: %s and %s (index %d).",
						selFrame.Name4(), name4, idxFrame))
				return // halt any further processing
			}

			idxsToDelete = append(idxsToDelete, idxFrame)
		}

		if !prompt.OkCancel(me.wnd, "Deleting frames", win.StrOptNone(),
			fmt.Sprintf("Are you sure you want to delete %d frame(s)?", len(idxsToDelete))) {
			return
		}

		tag.DeleteFrames(func(idx int, _ *id3v2.Frame) (willDelete bool) {
			for _, idxFrame := range idxsToDelete {
				if idx == idxFrame {
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
