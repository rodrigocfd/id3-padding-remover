package dlgfields

import (
	"id3fit/id3v2"
	"id3fit/timecount"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

type DlgFields struct {
	wnd            ui.WindowControl
	chkArtist      ui.CheckBox
	txtArtist      ui.Edit
	chkTitle       ui.CheckBox
	txtTitle       ui.Edit
	chkSubtitle    ui.CheckBox
	txtSubtitle    ui.Edit
	chkAlbum       ui.CheckBox
	txtAlbum       ui.Edit
	chkTrack       ui.CheckBox
	txtTrack       ui.Edit
	chkYear        ui.CheckBox
	txtYear        ui.Edit
	chkGenre       ui.CheckBox
	cmbGenre       ui.ComboBox
	chkComposer    ui.CheckBox
	txtComposer    ui.Edit
	chkOriginal    ui.CheckBox
	txtOriginal    ui.Edit
	chkPerformer   ui.CheckBox
	txtPerformer   ui.Edit
	chkComment     ui.CheckBox
	txtComment     ui.Edit
	btnClearChecks ui.Button
	btnSave        ui.Button

	onSaveCb   func(t0 timecount.TimeCount)
	tagsLoaded []*id3v2.Tag
	fields     []Field
}

type Field struct {
	Id  id3v2.TEXT
	Chk ui.CheckBox
	Txt ui.AnyNativeControl
}

func NewDlgFields(
	parent ui.AnyParent, position win.POINT,
	horz ui.HORZ, vert ui.VERT) *DlgFields {

	wnd := ui.NewWindowControlDlg(parent, DLG_MODAL, position, horz, vert)

	me := &DlgFields{
		wnd:            wnd,
		chkArtist:      ui.NewCheckBoxDlg(wnd, CHK_ARTIST, ui.HORZ_NONE, ui.VERT_NONE),
		txtArtist:      ui.NewEditDlg(wnd, TXT_ARTIST, ui.HORZ_NONE, ui.VERT_NONE),
		chkTitle:       ui.NewCheckBoxDlg(wnd, CHK_TITLE, ui.HORZ_NONE, ui.VERT_NONE),
		txtTitle:       ui.NewEditDlg(wnd, TXT_TITLE, ui.HORZ_NONE, ui.VERT_NONE),
		chkSubtitle:    ui.NewCheckBoxDlg(wnd, CHK_SUBTITLE, ui.HORZ_NONE, ui.VERT_NONE),
		txtSubtitle:    ui.NewEditDlg(wnd, TXT_SUBTITLE, ui.HORZ_NONE, ui.VERT_NONE),
		chkAlbum:       ui.NewCheckBoxDlg(wnd, CHK_ALBUM, ui.HORZ_NONE, ui.VERT_NONE),
		txtAlbum:       ui.NewEditDlg(wnd, TXT_ALBUM, ui.HORZ_NONE, ui.VERT_NONE),
		chkTrack:       ui.NewCheckBoxDlg(wnd, CHK_TRACK, ui.HORZ_NONE, ui.VERT_NONE),
		txtTrack:       ui.NewEditDlg(wnd, TXT_TRACK, ui.HORZ_NONE, ui.VERT_NONE),
		chkYear:        ui.NewCheckBoxDlg(wnd, CHK_YEAR, ui.HORZ_NONE, ui.VERT_NONE),
		txtYear:        ui.NewEditDlg(wnd, TXT_YEAR, ui.HORZ_NONE, ui.VERT_NONE),
		chkGenre:       ui.NewCheckBoxDlg(wnd, CHK_GENRE, ui.HORZ_NONE, ui.VERT_NONE),
		cmbGenre:       ui.NewComboBoxDlg(wnd, CMB_GENRE, ui.HORZ_NONE, ui.VERT_NONE),
		chkComposer:    ui.NewCheckBoxDlg(wnd, CHK_COMPOSER, ui.HORZ_NONE, ui.VERT_NONE),
		txtComposer:    ui.NewEditDlg(wnd, TXT_COMPOSER, ui.HORZ_NONE, ui.VERT_NONE),
		chkOriginal:    ui.NewCheckBoxDlg(wnd, CHK_ORIGINAL, ui.HORZ_NONE, ui.VERT_NONE),
		txtOriginal:    ui.NewEditDlg(wnd, TXT_ORIGINAL, ui.HORZ_NONE, ui.VERT_NONE),
		chkPerformer:   ui.NewCheckBoxDlg(wnd, CHK_PERFORMER, ui.HORZ_NONE, ui.VERT_NONE),
		txtPerformer:   ui.NewEditDlg(wnd, TXT_PERFORMER, ui.HORZ_NONE, ui.VERT_NONE),
		chkComment:     ui.NewCheckBoxDlg(wnd, CHK_COMMENT, ui.HORZ_NONE, ui.VERT_NONE),
		txtComment:     ui.NewEditDlg(wnd, TXT_COMMENT, ui.HORZ_NONE, ui.VERT_NONE),
		btnClearChecks: ui.NewButtonDlg(wnd, BTN_CLEARCHECKS, ui.HORZ_NONE, ui.VERT_NONE),
		btnSave:        ui.NewButtonDlg(wnd, BTN_SAVE, ui.HORZ_NONE, ui.VERT_NONE),
	}

	me.fields = []Field{
		{Id: id3v2.TEXT_ARTIST, Chk: me.chkArtist, Txt: me.txtArtist},
		{Id: id3v2.TEXT_TITLE, Chk: me.chkTitle, Txt: me.txtTitle},
		{Id: id3v2.TEXT_SUBTITLE, Chk: me.chkSubtitle, Txt: me.txtSubtitle},
		{Id: id3v2.TEXT_ALBUM, Chk: me.chkAlbum, Txt: me.txtAlbum},
		{Id: id3v2.TEXT_TRACK, Chk: me.chkTrack, Txt: me.txtTrack},
		{Id: id3v2.TEXT_YEAR, Chk: me.chkYear, Txt: me.txtYear},
		{Id: id3v2.TEXT_GENRE, Chk: me.chkGenre, Txt: me.cmbGenre},
		{Id: id3v2.TEXT_COMPOSER, Chk: me.chkComposer, Txt: me.txtComposer},
		{Id: id3v2.TEXT_ORIGINAL, Chk: me.chkOriginal, Txt: me.txtOriginal},
		{Id: id3v2.TEXT_PERFORMER, Chk: me.chkPerformer, Txt: me.txtPerformer},
		{Id: id3v2.TEXT_COMMENT, Chk: me.chkComment, Txt: me.txtComment},
	}

	me.eventsWm()
	return me
}
