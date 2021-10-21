package dlgfields

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/prompt"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgFields) eventsWm() {
	CHKS, INPS := me.checksAndInputs()

	me.wnd.On().WmInitDialog(func(_ wm.InitDialog) bool {
		genresPath := win.Path.ExePath() + "\\id3fit-genres.txt"
		if !win.Path.Exists(genresPath) {
			prompt.Error(me.wnd, "No genres file", nil,
				fmt.Sprintf("Genres file not found:\n\n%s", genresPath))
		} else {
			fin, _ := win.OpenFileMapped(genresPath, co.OPEN_FILE_READ_EXISTING)
			genres := fin.ReadLines()
			fin.Close()
			me.cmbGenre.Items().Add(genres...)
		}

		return true
	})

	for i, chk := range CHKS {
		func(i int, chk ui.CheckBox) {
			chk.On().BnClicked(func() {
				INPS[i].Hwnd().EnableWindow(chk.IsChecked()) // enable/disable input with checkbox

				atLeastOneEnabled := false
				for _, inp := range INPS {
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

		for i := 0; i < len(CHKS); i++ {
			if !CHKS[i].IsChecked() {
				continue
			}

			newText := INPS[i].Hwnd().GetWindowText()
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
