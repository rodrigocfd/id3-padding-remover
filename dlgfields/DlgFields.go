package dlgfields

import (
	"id3fit/id3v2"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

type DlgFields struct {
	wnd         ui.WindowControl
	chkArtist   ui.CheckBox
	txtArtist   ui.Edit
	chkTitle    ui.CheckBox
	txtTitle    ui.Edit
	chkAlbum    ui.CheckBox
	txtAlbum    ui.Edit
	chkTrack    ui.CheckBox
	txtTrack    ui.Edit
	chkYear     ui.CheckBox
	txtYear     ui.Edit
	chkGenre    ui.CheckBox
	cmbGenre    ui.ComboBox
	chkComposer ui.CheckBox
	txtComposer ui.Edit
	chkComment  ui.CheckBox
	txtComment  ui.Edit
	btnSave     ui.Button

	onSaveCb   func(t0 timecount.TimeCount)
	tagsLoaded []*id3v2.Tag
}

func NewDlgFields(
	parent ui.AnyParent, position win.POINT,
	horz ui.HORZ, vert ui.VERT) *DlgFields {

	wnd := ui.NewWindowControlDlg(parent, DLG_MODAL, position, horz, vert)

	me := &DlgFields{
		wnd:         wnd,
		chkArtist:   ui.NewCheckBoxDlg(wnd, CHK_ARTIST, ui.HORZ_NONE, ui.VERT_NONE),
		txtArtist:   ui.NewEditDlg(wnd, TXT_ARTIST, ui.HORZ_NONE, ui.VERT_NONE),
		chkTitle:    ui.NewCheckBoxDlg(wnd, CHK_TITLE, ui.HORZ_NONE, ui.VERT_NONE),
		txtTitle:    ui.NewEditDlg(wnd, TXT_TITLE, ui.HORZ_NONE, ui.VERT_NONE),
		chkAlbum:    ui.NewCheckBoxDlg(wnd, CHK_ALBUM, ui.HORZ_NONE, ui.VERT_NONE),
		txtAlbum:    ui.NewEditDlg(wnd, TXT_ALBUM, ui.HORZ_NONE, ui.VERT_NONE),
		chkTrack:    ui.NewCheckBoxDlg(wnd, CHK_TRACK, ui.HORZ_NONE, ui.VERT_NONE),
		txtTrack:    ui.NewEditDlg(wnd, TXT_TRACK, ui.HORZ_NONE, ui.VERT_NONE),
		chkYear:     ui.NewCheckBoxDlg(wnd, CHK_YEAR, ui.HORZ_NONE, ui.VERT_NONE),
		txtYear:     ui.NewEditDlg(wnd, TXT_YEAR, ui.HORZ_NONE, ui.VERT_NONE),
		chkGenre:    ui.NewCheckBoxDlg(wnd, CHK_GENRE, ui.HORZ_NONE, ui.VERT_NONE),
		cmbGenre:    ui.NewComboBoxDlg(wnd, CMB_GENRE, ui.HORZ_NONE, ui.VERT_NONE),
		chkComposer: ui.NewCheckBoxDlg(wnd, CHK_COMPOSER, ui.HORZ_NONE, ui.VERT_NONE),
		txtComposer: ui.NewEditDlg(wnd, TXT_COMPOSER, ui.HORZ_NONE, ui.VERT_NONE),
		chkComment:  ui.NewCheckBoxDlg(wnd, CHK_COMMENT, ui.HORZ_NONE, ui.VERT_NONE),
		txtComment:  ui.NewEditDlg(wnd, TXT_COMMENT, ui.HORZ_NONE, ui.VERT_NONE),
		btnSave:     ui.NewButtonDlg(wnd, BTN_SAVE, ui.HORZ_NONE, ui.VERT_NONE),
	}

	me.eventsWm()
	return me
}

func (me *DlgFields) OnSave(cb func(t0 timecount.TimeCount)) {
	me.onSaveCb = cb
}

func (me *DlgFields) Feed(tags []*id3v2.Tag) {
	chks, inps := me.checksAndInputs()
	for _, chk := range chks {
		chk.Hwnd().EnableWindow(len(tags) > 0) // if zero MP3s selected, disable checkboxes
	}

	if len(tags) == 0 { // zero MP3s selected
		for c := 0; c < len(chks); c++ {
			inps[c].Hwnd().SetWindowText("")
			chks[c].SetCheckStateAndTrigger(co.BST_UNCHECKED)
		}
	} else {
		names4 := id3v2.TextFieldConsts()

		for n := 0; n < len(names4); n++ {
			if firstText, ok := tags[0].TextByName4(names4[n]); ok {
				sameStr := true

				for t := 1; t < len(tags); t++ { // subsequent tags
					if otherText, ok := tags[t].TextByName4(names4[n]); ok {
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
					inps[n].Hwnd().SetWindowText(firstText)
					chks[n].SetCheckStateAndTrigger(co.BST_CHECKED)
				} else {
					inps[n].Hwnd().SetWindowText("")
					chks[n].SetCheckStateAndTrigger(co.BST_UNCHECKED)
				}

			} else { // frame absent in first tag
				inps[n].Hwnd().SetWindowText("")
				chks[n].SetCheckStateAndTrigger(co.BST_UNCHECKED)
			}
		}
	}

	me.tagsLoaded = tags
}

func (me *DlgFields) checksAndInputs() (chks []ui.CheckBox, inps []ui.AnyNativeControl) {
	// Note: This must be in sync with id3v2.TextFieldConsts().
	chks = []ui.CheckBox{me.chkArtist, me.chkTitle, me.chkAlbum,
		me.chkTrack, me.chkYear, me.chkGenre, me.chkComposer, me.chkComment}
	inps = []ui.AnyNativeControl{me.txtArtist, me.txtTitle, me.txtAlbum,
		me.txtTrack, me.txtYear, me.cmbGenre, me.txtComposer, me.txtComment}
	return
}
