package dlgfields

import (
	"id3fit/id3v2"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgFields) OnSave(cb func(t0 timecount.TimeCount)) {
	me.onSaveCb = cb
}

func (me *DlgFields) Feed(tags []*id3v2.Tag) {
	for _, field := range me.fields {
		field.Chk.Hwnd().EnableWindow(len(tags) > 0) // if zero MP3s selected, disable checkboxes
	}

	if len(tags) == 0 { // zero MP3s selected
		for _, field := range me.fields {
			field.Chk.SetCheckState(co.BST_UNCHECKED)
			field.Txt.SetText("")
			field.Txt.Hwnd().EnableWindow(false)
		}
	} else {
		for _, field := range me.fields {
			if text, same := id3v2.TagSameValueAcrossAll(tags, field.FrameId); same {
				field.Chk.SetCheckState(co.BST_CHECKED)
				field.Txt.SetText(text)
				field.Txt.Hwnd().EnableWindow(true)
			} else {
				field.Chk.SetCheckState(co.BST_UNCHECKED)
				field.Txt.SetText("")
				field.Txt.Hwnd().EnableWindow(false)
			}
		}
	}

	me.tagsLoaded = tags
	me.enableButtonsIfAtLeastOneChecked()
}

func (me *DlgFields) enableButtonsIfAtLeastOneChecked() {
	atLeastOneChecked := false
	for _, otherField := range me.fields {
		if otherField.Txt.Hwnd().IsWindowEnabled() {
			atLeastOneChecked = true
			break
		}
	}
	me.btnClearChecks.Hwnd().EnableWindow(atLeastOneChecked)
	me.btnSave.Hwnd().EnableWindow(atLeastOneChecked)
}
