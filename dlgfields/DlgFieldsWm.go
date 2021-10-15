package dlgfields

import (
	"id3fit/id3v2"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui"
)

func (me *DlgFields) eventsWm() {
	chks, inps := me.checksAndInputs()

	for i, chk := range chks {
		func(i int, chk ui.CheckBox) {
			chk.On().BnClicked(func() {
				inps[i].Hwnd().EnableWindow(chk.IsChecked()) // enable/disable input with checkbox

				atLeastOneEnabled := false
				for _, inp := range inps {
					if inp.Hwnd().IsWindowEnabled() {
						atLeastOneEnabled = true
						break
					}
				}
				me.btnSave.Hwnd().EnableWindow(atLeastOneEnabled)
			})
		}(i, chk)
	}

	me.btnSave.On().BnClicked(func() {
		t0 := timecount.New()
		fields := id3v2.TextFieldConsts()

		for i := 0; i < len(chks); i++ {
			if !chks[i].IsChecked() {
				continue
			}

			newText := inps[i].Hwnd().GetWindowText()
			for _, tag := range me.tagsLoaded {
				// Empty text will delete the frame.
				// Tags are not flushed to disk here, it's DlgMain's job.
				tag.SetTextByName4(fields[i], newText)
			}
		}

		if me.onSaveCb != nil {
			me.onSaveCb(t0)
		}
	})
}
