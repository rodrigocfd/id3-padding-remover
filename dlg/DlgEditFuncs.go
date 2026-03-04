//go:build windows

package dlg

import (
	_ "embed"
	"fmt"
	"id3fit/id3v2"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/wstr"
)

//go:embed genres.txt
var genres string

func (me *DlgEdit) fillComboGenres() {
	for _, input := range me.inputs {
		if input.txt.CtrlId() == CMB_GENRE {
			cmb, _ := input.txt.(*ui.ComboBox)
			for _, genre := range wstr.SplitLines(genres) {
				if genre != "" {
					cmb.AddItem(genre)
				}
			}
			break
		}
	}
}

func (me *DlgEdit) fillTextboxes() {
	for _, input := range me.inputs {
		name4 := FIELD_NAMES[input.chk.CtrlId()]
		var pFrame *id3v2.Frame

		if len(me.tags) == 1 {
			pFrame = me.tags[0].FrameByName4(name4)
		} else {
			pFrame = id3v2.SameFrameAcrossAllTags(name4, me.tags)
		}

		if pFrame == nil { // if we don't have the same value across all files, show nothing
			input.chk.SetCheckAndTrigger(false)
			input.txt.Hwnd().SetWindowText("")
		} else {
			input.chk.SetCheckAndTrigger(true)
			input.txt.Hwnd().SetWindowText(pFrame.Body().AsText())
		}
	}
}

// Fills the right-placed list view with all existing frames, if editing 1 tag.
func (me *DlgEdit) fillFramesList() {
	if len(me.tags) == 1 { // editing 1 tag
		me.lstFrames.DeleteAllItems() // first clean, then render
		for _, pFrame := range me.tags[0].Frames() {
			me.lstFrames.AddItem(pFrame.Name4(), pFrame.Body().AsText())
		}
		me.lstFrames.Col(1).SetWidthToFill()
	} else { // editing multiple tags
		me.lstFrames.Hwnd().EnableWindow(false)
		me.lstFrames.AddItem("", fmt.Sprintf("%d files...", len(me.tags)))
	}
}

// Displays resolution and size of the cover art in the label.
func (me *DlgEdit) fillPicInfo(pixels win.SIZE, nBytes int) {
	if nBytes == 0 {
		if len(me.tags) == 1 {
			me.lblPic.Hwnd().SetWindowText("(no picture)")
		} else {
			me.lblPic.Hwnd().SetWindowText("(different pictures)")
		}
	} else {
		me.chkPick.SetCheck(true)
		me.lblPic.Hwnd().SetWindowText(fmt.Sprintf("%dx%d px, %s",
			pixels.Cx, pixels.Cy, wstr.FmtBytes(nBytes)))
	}
}

func (me *DlgEdit) writeTextsToTags() {
	for _, input := range me.inputs {
		if !input.chk.IsChecked() {
			continue // skip unchecked fields
		}

		text, _ := input.txt.Hwnd().GetWindowText()
		text = strings.TrimSpace(text)

		for _, pTag := range me.tags {
			name4 := FIELD_NAMES[input.chk.CtrlId()]
			if pFrame := pTag.FrameByName4(name4); pFrame != nil { // the frame already exists in this tag
				if text == "" { // empty text will remove the frame
					pTag.RemoveFrameIf(func(pFrame *id3v2.Frame) bool {
						return pFrame.Name4() == name4 // note: with TXXX, will remove all TXXX
					})
				} else { // otherwise update the text in the frame
					pFrame.Body().ForceText(text)
				}
			} else { // the frame doesn't exist in this tag yet; create it
				if text != "" {
					pTag.AddFrameWithText(name4, text)
				}
			}
		}
	}
}
