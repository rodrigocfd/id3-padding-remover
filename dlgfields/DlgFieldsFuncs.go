package dlgfields

import (
	"id3fit/id3v2"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/win/co"
)

// Stores the callback to be called when the user clicks the Save button.
// This is called after the tag slice is updated.
func (me *DlgFields) OnSave(cb func(t0 timecount.TimeCount)) {
	me.onSaveCb = cb
}

// Puts the contents of the multiple tags into the fields.
func (me *DlgFields) Feed(selectedTags []*id3v2.Tag) {
	for _, field := range me.fields {
		field.Chk.Hwnd().EnableWindow(len(selectedTags) > 0) // if zero MP3s selected, disable checkboxes
	}

	if len(selectedTags) == 0 { // zero MP3s selected
		for _, field := range me.fields {
			field.Chk.SetCheckState(co.BST_UNCHECKED)
			field.Txt.SetText("")
			field.Txt.Hwnd().EnableWindow(false)
		}
	} else {
		for _, field := range me.fields {
			if text, same := id3v2.TagSameValueAcrossAll(selectedTags, field.FrameId); same {
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

	me.selectedTags = selectedTags // keep the tags slice, will be updated when user clicks Save
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
