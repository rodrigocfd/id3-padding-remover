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
			field.Txt.Hwnd().SetWindowText("")
			field.Chk.SetCheckStateAndTrigger(co.BST_UNCHECKED)
		}
	} else {
		for _, field := range me.fields {
			if firstText, ok := tags[0].TextByName4(field.Id); ok {
				sameStr := true // the field value is the same across all tags?

				for t := 1; t < len(tags); t++ { // subsequent tags
					if otherText, ok := tags[t].TextByName4(field.Id); ok {
						if otherText != firstText {
							sameStr = false
							break
						}
					} else { // frame absent in subsequent tag
						sameStr = false
						break
					}
				}

				if sameStr {
					field.Txt.Hwnd().SetWindowText(firstText)
					field.Chk.SetCheckStateAndTrigger(co.BST_CHECKED)
				} else {
					field.Txt.Hwnd().SetWindowText("")
					field.Chk.SetCheckStateAndTrigger(co.BST_UNCHECKED)
				}

			} else { // frame absent in first tag
				field.Txt.Hwnd().SetWindowText("")
				field.Chk.SetCheckStateAndTrigger(co.BST_UNCHECKED)
			}
		}
	}

	me.tagsLoaded = tags
}
