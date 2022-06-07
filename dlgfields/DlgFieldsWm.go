package dlgfields

import (
	"fmt"
	"id3fit/id3v2"
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
			ui.TaskDlg.Error(me.wnd, "No genres file", win.StrOptNone(),
				fmt.Sprintf("Genres file not found:\n\n%s", genresTxt))
		} else {
			genres := func() []string {
				fin, _ := win.FileMappedOpen(genresTxt, co.FILE_OPEN_READ_EXISTING)
				defer fin.Close()
				return fin.ReadLines()
			}()

			for i := range me.fields {
				field := &me.fields[i]
				if field.FrameId == id3v2.FRAMETXT_GENRE { // find the genre combobox
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
				if field.Chk.IsChecked() {
					field.Txt.Focus() // if checkbox was checked, focus the edit
				}
				me.enableButtonsIfAtLeastOneChecked()
			})

		}(&me.fields[i])
	}

	me.btnClearChecks.On().BnClicked(func() {
		for _, field := range me.fields {
			field.Chk.SetCheckState(co.BST_UNCHECKED)
			field.Txt.Hwnd().EnableWindow(false)
		}
		me.enableButtonsIfAtLeastOneChecked()
	})

	me.btnSave.On().BnClicked(func() {
		t0 := timecount.New()

		for _, field := range me.fields {
			// Checked fields are be updated.
			// Unchecked ones remain untouched.
			if field.Chk.IsChecked() {
				newText := strings.TrimSpace(field.Txt.Text())
				for _, tag := range me.tagsLoaded {
					// Empty text will delete the frame.
					// Tags are changed but not flushed to disk here, it's DlgMain's job.
					tag.SetTextByFrameId(field.FrameId, newText)
				}
			}
		}

		if me.onSaveCb != nil {
			me.onSaveCb(t0) // invoke parent's callback
		}
	})
}
