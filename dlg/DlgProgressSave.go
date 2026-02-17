//go:build windows

package dlg

import (
	"fmt"
	"strings"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
)

// Modal dialog to load ID3v2 tags from MP3 files.
type DlgProgressSave struct {
	wnd        *ui.Modal
	prog       *ui.ProgressBar
	tagsToSave []TagAndPath
}

// Constructor; blocks until the modal is closed.
func ShowDlgProgressSave(parent ui.Parent, outgoingTags []TagAndPath) {
	wnd := ui.NewModalDlg(parent, DLG_PROGRESS)
	prog := ui.NewProgressBarDlg(wnd, PRO_PRO, ui.LAY_HOLD_HOLD)

	me := &DlgProgressSave{wnd, prog, outgoingTags}
	me.events()
	me.wnd.ShowModal() // blocks until the modal is closed
}

func (me *DlgProgressSave) events() {
	me.wnd.On().WmInitDialog(func(_ ui.WmInitDialog) bool {
		me.prog.SetRange(0, len(me.tagsToSave))
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("0/%d file(s) written...", len(me.tagsToSave)))
		go me.saveAsync()
		return true
	})
}

func (me *DlgProgressSave) saveAsync() {
	type Failure struct { // a register of one saving error
		path string
		err  error
	}
	failures := []Failure{} // to store all failures we get

	for idxTag, tagToSave := range me.tagsToSave { // process each tag sequentially
		if err := tagToSave.Tag.SaveToFile(tagToSave.Path); err != nil {
			failures = append(failures, Failure{tagToSave.Path, err}) // store error, and keep going
		}

		me.wnd.UiThread(func() { // UI progress feedback
			me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%d/%d file(s) written...", idxTag+1, len(me.tagsToSave)))
			me.prog.SetPos(idxTag + 1)
		})
	}

	me.wnd.UiThread(func() { // UI final feedback
		if len(failures) > 0 { // any errors?
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%d file(s) failed to save:", len(failures)))
			for _, fail := range failures {
				sb.WriteString("\n\n")
				sb.WriteString(fail.path)
				sb.WriteString("\n")
				sb.WriteString(fail.err.Error())
			}
			ui.MsgError(me.wnd, "Error saving file(s)", "", sb.String())
		}
		me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
	})
}
