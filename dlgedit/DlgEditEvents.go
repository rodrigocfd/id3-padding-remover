//go:build windows

package dlgedit

import (
	"fmt"
	"id3fit/ids"
	"slices"
	"strings"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
)

func (me *DlgEdit) events() {

	me.wnd.On().WmInitDialog(func(_ ui.WmInitDialog) bool {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("Editing %d ID3v2 tag(s)", len(me.tags)))

		me.lstFrames.SetExtendedStyle(true, co.LVS_EX_FULLROWSELECT|co.LVS_EX_GRIDLINES)
		me.lstFrames.Cols.Add("Frame", ui.DpiX(56))
		me.lstFrames.Cols.Add("Value", ui.DpiX(100))

		me.fillComboGenres()
		me.fillTextboxes()
		me.fillFramesList()
		pixels, nBytes := me.wndPic.LoadPicture(me.tags)
		me.fillPicInfo(pixels, nBytes)

		return true
	})

	me.wnd.On().WmInitMenuPopup(func(p ui.WmInitMenuPopup) {
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
				input.txt.Hwnd().EnableWindow(true)
				input.txt.Focus()
			} else {
				input.txt.Hwnd().EnableWindow(false)
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
			txt, _ := input.txt.Hwnd().GetWindowText()
			if strings.TrimSpace(txt) != "" {
				input.chk.SetCheckAndTrigger(true) // check if has text
			}
		}
	})

	me.wnd.On().WmCommandAccelMenu(uint16(co.ID_OK), func() {
		me.writeTextsToTags()
		me.result = co.ID_OK
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
	})

	me.wnd.On().WmCommandAccelMenu(uint16(co.ID_CANCEL), func() {
		me.result = co.ID_CANCEL
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FRAMES_MOVEUP, func() {
		focusedItem, hasFocused := me.lstFrames.Items.Focused()

		selItems := me.lstFrames.Items.Selected()
		for _, sel := range selItems {
			idx := sel.Index()
			me.tags[0].Frames()[idx], me.tags[0].Frames()[idx-1] =
				me.tags[0].Frames()[idx-1], me.tags[0].Frames()[idx]
		}
		me.fillFramesList()
		for _, sel := range selItems {
			prev, _ := sel.Prev()
			prev.Select(true)
		}

		if hasFocused {
			me.lstFrames.Items.Get(focusedItem.Index() - 1).Focus()
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FRAMES_MOVEDOWN, func() {
		focusedItem, hasFocused := me.lstFrames.Items.Focused()

		selItems := me.lstFrames.Items.Selected()
		for _, selItem := range slices.Backward(selItems) {
			idx := selItem.Index()
			me.tags[0].Frames()[idx], me.tags[0].Frames()[idx+1] =
				me.tags[0].Frames()[idx+1], me.tags[0].Frames()[idx]
		}
		me.fillFramesList()
		for _, selItem := range selItems {
			next, _ := selItem.Next()
			next.Select(true)
		}

		if hasFocused {
			me.lstFrames.Items.Get(focusedItem.Index() + 1).Focus()
		}
	})

	me.wnd.On().WmCommandAccelMenu(ids.MNU_FRAMES_DELETE, func() {
		selItems := me.lstFrames.Items.Selected()
		text := fmt.Sprintf("Do you want to remove %d frame(s)?", len(selItems))

		if ui.MsgOkCancel(me.wnd, "Remove frames", "", text, "&Remove") {
			for _, selItem := range slices.Backward(selItems) {
				me.tags[0].RemoveFrame(selItem.Index())
			}
			me.fillFramesList()
		}
	})

}
