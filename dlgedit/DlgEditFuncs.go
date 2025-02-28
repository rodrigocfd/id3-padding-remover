//go:build windows

package dlgedit

import (
	_ "embed"
	"errors"
	"fmt"
	"id3fit/id3v2"
	"id3fit/ids"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/ole"
)

func (me *DlgEdit) loadComPicture() {
	if apic := id3v2.SameFrameAcrossAllTags("APIC", me.tags); apic != nil {
		body, ok := apic.Body().(*id3v2.BodyPicture)
		if !ok {
			me.loadComPictureError("Invalid APIC frame",
				errors.New("APIC frame does not contain BodyPicture body type."))
		}
		stream, err := ole.SHCreateMemStream(&me.comRel, body.Bin)
		if err != nil {
			me.loadComPictureError("Error creating stream", err)
		}
		me.pic, err = ole.OleLoadPicture(&me.comRel, &stream, uint(len(body.Bin)), true)
		if err != nil {
			me.loadComPictureError("Error parsing stream", err)
		}
	}
}

func (me *DlgEdit) loadComPictureError(caption string, err error) {
	win.TaskDialogIndirect(win.TASKDIALOGCONFIG{
		HwndParent:      me.wnd.Hwnd(),
		WindowTitle:     "Picture error",
		MainInstruction: caption,
		Content:         err.Error(),
		HMainIcon:       win.TdcIconTdi(co.TDICON_ERROR),
		CommonButtons:   co.TDCBF_OK,
		Flags:           co.TDF_ALLOW_DIALOG_CANCELLATION | co.TDF_POSITION_RELATIVE_TO_WINDOW,
	})
}

func (me *DlgEdit) updateTitlebar() {
	caption := "Edit ID3v2 tags - "
	if len(me.tags) == 1 {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%s %d frames", caption, len(me.orderedFrames)))
	} else {
		me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%s %d files", caption, len(me.tags)))
	}
}

//go:embed genres.txt
var genres string

func (me *DlgEdit) fillComboGenres() {
	for _, input := range me.inputs {
		if input.in.CtrlId() == ids.CMB_GENRE {
			cmb, _ := input.in.(*ui.ComboBox)
			for genre := range win.Str.IterLines(genres) {
				if genre != "" {
					cmb.Items.Add(genre)
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
			input.in.Hwnd().SetWindowText("")
		} else {
			input.chk.SetCheckAndTrigger(true)
			input.in.Hwnd().SetWindowText(pFrame.Body().AsText())
		}
	}
}

func (me *DlgEdit) showPicSize() {
	if me.pic.Ppvt() != nil {
		hdcScreen, _ := win.HWND(0).GetDC()
		defer win.HWND(0).ReleaseDC(hdcScreen)
		szPic, _ := me.pic.SizePixels(hdcScreen)
		me.lblPic.Hwnd().SetWindowText(fmt.Sprintf("%d x % d pixels", szPic.Cx, szPic.Cy))
	} else if len(me.tags) == 1 {
		me.lblPic.Hwnd().SetWindowText("(no picture)")
	} else {
		me.lblPic.Hwnd().SetWindowText("(different pictures)")
	}
}

func (me *DlgEdit) fillFramesList() {
	if len(me.tags) == 1 {
		me.lstFrames.Items.DeleteAll() // first clean, then render
		for _, pFrame := range me.orderedFrames {
			me.lstFrames.Items.Add(pFrame.Name4(), pFrame.Body().AsText())
		}
		me.lstFrames.Cols.Get(1).SetWidthToFill()
	} else {
		me.lstFrames.Hwnd().EnableWindow(false)
		me.lstFrames.Items.Add("", fmt.Sprintf("%d files...", len(me.tags)))
	}
}

func (me *DlgEdit) writeTextsToTags() {
	if len(me.tags) == 1 {
		me.tags[0].ReplaceFrames(me.orderedFrames)
	}

	for _, input := range me.inputs {
		if !input.chk.IsChecked() {
			continue // skip unchecked fields
		}

		text, _ := input.in.Hwnd().GetWindowText()
		text = strings.TrimSpace(text)

		for _, pTag := range me.tags {
			name4 := FIELD_NAMES[input.chk.CtrlId()]
			if pFrame := pTag.FrameByName4(name4); pFrame != nil { // the frame already exists in this tag
				if text == "" { // empty text will remove the frame
					pTag.RemoveFrameIf(func(pFrame *id3v2.Frame) bool {
						return pFrame.Name4() == name4 // note: with TXXX, will remove all TXXX
					})
				} else {
					pFrame.Body().ForceText(text)
				}
			} else { // the frame doesn't exist in this tag yet
				if text != "" {
					pTag.AddFrameWithText(name4, text)
				}
			}
		}
	}

	// The file writing itself is made by DlgMain.
}
