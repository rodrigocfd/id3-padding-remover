package dlgfields

import (
	"fmt"
	"id3fit/id3v2"
	"id3fit/prompt"
	"id3fit/timecount"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func (me *DlgFields) eventsWm() {
	me.wnd.On().WmInitDialog(func(_ wm.InitDialog) bool {
		if genresTxt := win.Path.ExePath() + "\\id3fit-genres.txt"; !win.Path.Exists(genresTxt) {
			prompt.Error(me.wnd, "No genres file", nil,
				fmt.Sprintf("Genres file not found:\n\n%s", genresTxt))
		} else {
			genres := func() []string {
				fin, _ := win.OpenFileMapped(genresTxt, co.OPEN_FILE_READ_EXISTING)
				defer fin.Close()
				return fin.ReadLines()
			}()

			for i := range me.fields {
				field := &me.fields[i]
				if field.FrameId == id3v2.TEXT_GENRE {
					cmbGenre := field.Txt.(ui.ComboBox)
					cmbGenre.Items().Add(genres...)
					break
				}
			}
		}

		return true
	})

	for i := range me.fields {
		func(field *Field) {

			field.Chk.On().BnClicked(func() {
				field.Txt.Hwnd().EnableWindow(field.Chk.IsChecked()) // enable/disable input with checkbox

				atLeastOneEnabled := false
				for _, otherField := range me.fields {
					if otherField.Txt.Hwnd().IsWindowEnabled() {
						atLeastOneEnabled = true
						break
					}
				}
				me.btnClearChecks.Hwnd().EnableWindow(atLeastOneEnabled)
				me.btnSave.Hwnd().EnableWindow(atLeastOneEnabled)
			})

		}(&me.fields[i])
	}

	me.btnClearChecks.On().BnClicked(func() {
		for _, field := range me.fields {
			field.Chk.SetCheckStateAndTrigger(co.BST_UNCHECKED)
		}
	})

	me.btnSave.On().BnClicked(func() {
		t0 := timecount.New()

		for _, field := range me.fields {
			if !field.Chk.IsChecked() {
				continue
			}

			newText := strings.TrimSpace(field.Txt.Hwnd().GetWindowText())
			for _, tag := range me.tagsLoaded {
				// Empty text will delete the frame.
				// Tags are not flushed to disk here, it's DlgMain's job.
				tag.SetTextByName4(field.FrameId, newText)
			}
		}

		if me.onSaveCb != nil {
			me.onSaveCb(t0)
		}
	})
}
