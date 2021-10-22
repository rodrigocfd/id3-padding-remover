package dlgfields

import (
	"fmt"
	"id3fit/prompt"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgFields) eventsWm() {
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

	for _, field := range me.fields {
		func(field Field) {

			field.Chk.On().BnClicked(func() {
				field.Txt.Hwnd().EnableWindow(field.Chk.IsChecked()) // enable/disable input with checkbox

				atLeastOneEnabled := false
				for _, otherField := range me.fields {
					if otherField.Txt.Hwnd().IsWindowEnabled() {
						atLeastOneEnabled = true
						break
					}
				}
				me.btnSave.Hwnd().EnableWindow(atLeastOneEnabled)
			})

		}(field)
	}

	me.btnSave.On().BnClicked(func() {
		t0 := timecount.New()

		for _, field := range me.fields {
			if !field.Chk.IsChecked() {
				continue
			}

			newText := field.Txt.Hwnd().GetWindowText()
			for _, tag := range me.tagsLoaded {
				// Empty text will delete the frame.
				// Tags are not flushed to disk here, it's DlgMain's job.
				tag.SetTextByName4(field.Id, newText)
			}
		}

		if me.onSaveCb != nil {
			me.onSaveCb(t0)
		}
	})
}
