package dlgfields

import (
	"id3fit/id3v2"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

type Field struct {
	FrameId id3v2.TEXT
	ChkId   int
	TxtId   int
	Chk     ui.CheckBox
	Txt     ui.AnyTextControl
}

type DlgFields struct {
	wnd            ui.WindowControl
	fields         []Field
	btnClearChecks ui.Button
	btnSave        ui.Button

	onSaveCb   func(t0 timecount.TimeCount)
	tagsLoaded []*id3v2.Tag
}

func NewDlgFields(
	parent ui.AnyParent,
	position win.POINT,
	horz ui.HORZ, vert ui.VERT) *DlgFields {

	wnd := ui.NewWindowControlDlg(parent, DLG_MODAL, position, horz, vert)

	me := &DlgFields{
		wnd: wnd,
		fields: []Field{
			{FrameId: id3v2.TEXT_ARTIST, ChkId: CHK_ARTIST, TxtId: TXT_ARTIST},
			{FrameId: id3v2.TEXT_TITLE, ChkId: CHK_TITLE, TxtId: TXT_TITLE},
			{FrameId: id3v2.TEXT_SUBTITLE, ChkId: CHK_SUBTITLE, TxtId: TXT_SUBTITLE},
			{FrameId: id3v2.TEXT_ALBUM, ChkId: CHK_ALBUM, TxtId: TXT_ALBUM},
			{FrameId: id3v2.TEXT_TRACK, ChkId: CHK_TRACK, TxtId: TXT_TRACK},
			{FrameId: id3v2.TEXT_YEAR, ChkId: CHK_YEAR, TxtId: TXT_YEAR},
			{FrameId: id3v2.TEXT_GENRE, ChkId: CHK_GENRE, TxtId: CMB_GENRE},
			{FrameId: id3v2.TEXT_COMPOSER, ChkId: CHK_COMPOSER, TxtId: TXT_COMPOSER},
			{FrameId: id3v2.TEXT_LYRICIST, ChkId: CHK_LYRICIST, TxtId: TXT_LYRICIST},
			{FrameId: id3v2.TEXT_ORIGINAL, ChkId: CHK_ORIGINAL, TxtId: TXT_ORIGINAL},
			{FrameId: id3v2.TEXT_PERFORMER, ChkId: CHK_PERFORMER, TxtId: TXT_PERFORMER},
			{FrameId: id3v2.TEXT_COMMENT, ChkId: CHK_COMMENT, TxtId: TXT_COMMENT},
		},
		btnClearChecks: ui.NewButtonDlg(wnd, BTN_CLEARCHECKS, ui.HORZ_NONE, ui.VERT_NONE),
		btnSave:        ui.NewButtonDlg(wnd, BTN_SAVE, ui.HORZ_NONE, ui.VERT_NONE),
	}

	for i := range me.fields {
		field := &me.fields[i]
		field.Chk = ui.NewCheckBoxDlg(wnd, field.ChkId, ui.HORZ_NONE, ui.VERT_NONE)
		if field.FrameId == id3v2.TEXT_GENRE {
			field.Txt = ui.NewComboBoxDlg(wnd, field.TxtId, ui.HORZ_NONE, ui.VERT_NONE)
		} else {
			field.Txt = ui.NewEditDlg(wnd, field.TxtId, ui.HORZ_NONE, ui.VERT_NONE)
		}
	}

	me.eventsWm()
	return me
}
