//go:build windows

package dlgedit

import (
	"fmt"
	"id3fit/ids"
	"slices"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgEdit) events() {

	me.wnd.On().WmInitDialog(func(_ wm.InitDialog) bool {
		me.lstFrames.SetExtendedStyle(true, co.LVS_EX_FULLROWSELECT|co.LVS_EX_GRIDLINES)
		me.lstFrames.Cols.Add("Frame", ui.DpiX(56))
		me.lstFrames.Cols.Add("Value", ui.DpiX(100))

		me.updateTitlebar()
		me.fillComboGenres()
		me.fillTextboxes()
		me.showPicSize()
		me.fillFramesList()
		return true
	})

	me.wnd.On().WmInitMenuPopup(func(p wm.InitMenuPopup) {
		firstId, _ := p.HMenu().GetMenuItemID(0)
		if firstId == ids.MNU_FRAMES_MOVEUP {
			oneTag := len(me.tags) == 1
			hasSel := me.lstFrames.Items.SelectedCount() > 0
			firstItem := me.lstFrames.Items.Get(0)
			lastItem := me.lstFrames.Items.Last()

			p.HMenu().EnableMenuItemByCmd(oneTag && hasSel && !firstItem.IsSelected(), ids.MNU_FRAMES_MOVEUP)
			p.HMenu().EnableMenuItemByCmd(oneTag && hasSel && !lastItem.IsSelected(), ids.MNU_FRAMES_MOVEDOWN)
			p.HMenu().EnableMenuItemByCmd(oneTag && hasSel, ids.MNU_FRAMES_DELETE)
		}
	})

	for _, input := range me.inputs {
		input.chk.On().BnClicked(func() { // checkboxes enable/disable inputs
			if input.chk.IsChecked() {
				input.in.Hwnd().EnableWindow(true)
				input.in.Focus()
			} else {
				input.in.Hwnd().EnableWindow(false)
			}
		})
	}

	me.btnUncheckAll.On().BnClicked(func() {
		for _, input := range me.inputs {
			input.chk.SetCheckAndTrigger(false) // uncheck everyone
		}
	})

	me.btnCheckFilled.On().BnClicked(func() {
		for _, input := range me.inputs {
			txt, _ := input.in.Hwnd().GetWindowText()
			if strings.TrimSpace(txt) != "" {
				input.chk.SetCheckAndTrigger(true) // check if has text
			}
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FRAMES_MOVEUP, func() {
		selItems := slices.Collect(me.lstFrames.Items.IterSelected())
		for _, sel := range selItems {
			idx := sel.Index()
			me.orderedFrames[idx], me.orderedFrames[idx-1] =
				me.orderedFrames[idx-1], me.orderedFrames[idx]
		}
		me.fillFramesList()
		for _, sel := range selItems {
			prev, _ := sel.Prev()
			prev.Select(true)
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FRAMES_MOVEDOWN, func() {
		selItems := slices.Collect(me.lstFrames.Items.IterSelected())
		for _, selItem := range slices.Backward(selItems) {
			idx := selItem.Index()
			me.orderedFrames[idx], me.orderedFrames[idx+1] =
				me.orderedFrames[idx+1], me.orderedFrames[idx]
		}
		me.fillFramesList()
		for _, selItem := range selItems {
			next, _ := selItem.Next()
			next.Select(true)
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FRAMES_DELETE, func() {
		selItems := slices.Collect(me.lstFrames.Items.IterSelected())

		ret, _ := win.TaskDialogIndirect(win.TASKDIALOGCONFIG{
			HwndParent:  me.wnd.Hwnd(),
			WindowTitle: "Remove frames",
			Content:     fmt.Sprintf("Do you want to remove %d frame(s)?", len(selItems)),
			HMainIcon:   win.TdcIconTdi(co.TDICON_WARNING),
			Flags:       co.TDF_ALLOW_DIALOG_CANCELLATION | co.TDF_POSITION_RELATIVE_TO_WINDOW,
			Buttons: []win.TASKDIALOG_BUTTON{
				{Id: co.ID_OK, Text: "&Remove"},
				{Id: co.ID_CANCEL, Text: "&Cancel"},
			},
		})
		if ret == co.ID_OK {
			for _, selItem := range slices.Backward(selItems) {
				me.orderedFrames = slices.Delete(me.orderedFrames,
					selItem.Index(), selItem.Index()+1)
			}
			me.updateTitlebar()
			me.fillFramesList()
		}
	})

	me.wnd.On().WmCommandAccelMenu(uint16(co.ID_OK), func() {
		me.writeTextsToTags()
		me.result = co.ID_OK
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
	})

	me.wnd.On().WmCommandAccelMenu(uint16(co.ID_CANCEL), func() {
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
	})

}
