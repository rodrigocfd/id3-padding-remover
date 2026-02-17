//go:build windows

package dlg

import (
	"id3fit/id3v2"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
)

// Modal dialog to edit the fields of one or many tags.
type DlgEdit struct {
	wnd            *ui.Modal
	inputs         []CheckInput // checkbox + input fields
	chkPick        *ui.CheckBox
	wndPic         *WndPicture
	lblPic         *ui.Static
	lstFrames      *ui.ListView
	btnUncheckAll  *ui.Button
	btnCheckFilled *ui.Button

	tags   []*id3v2.Tag // cloned tags to be modified
	result co.ID        // returned when the user closes the modal
}

// A pair of a checkbox + input fields.
type CheckInput struct {
	chk *ui.CheckBox
	txt ui.ChildControl // Edit or ComboBox
}

// Constructor; blocks until the modal is closed.
func ShowDlgEdit(parent ui.Parent, tags []*id3v2.Tag) co.ID {
	wnd := ui.NewModalDlg(parent, DLG_EDIT)

	inputs := make([]CheckInput, 0, (TXT_COMMENT-CHK_ARTIST+1)/2)
	for id := CHK_ARTIST; id <= TXT_COMMENT; id += 2 {
		chk := ui.NewCheckBoxDlg(wnd, id, ui.LAY_HOLD_HOLD)
		var inp ui.ChildControl
		if id == CHK_GENRE {
			inp = ui.NewComboBoxDlg(wnd, CMB_GENRE, ui.LAY_HOLD_HOLD)
		} else {
			inp = ui.NewEditDlg(wnd, id+1, ui.LAY_HOLD_HOLD)
		}
		inputs = append(inputs, CheckInput{chk, inp})
	}

	chkPic := ui.NewCheckBoxDlg(wnd, CHK_PICTURE, ui.LAY_HOLD_HOLD)
	wndPic := NewWndPicture(wnd, ui.DpiX(420), ui.DpiY(60), ui.DpiX(200), ui.DpiY(200))
	lblPic := ui.NewStaticDlg(wnd, LBL_IMAGE_DESCR, ui.LAY_HOLD_HOLD)

	lstFrames := ui.NewListViewDlg(wnd, LST_FRAMES, MNU_FRAMES, ui.LAY_HOLD_HOLD)
	btnUncheckAll := ui.NewButtonDlg(wnd, BTN_UNCHECK_ALL, ui.LAY_HOLD_HOLD)
	btnCheckFilled := ui.NewButtonDlg(wnd, BTN_CHECK_FILLED, ui.LAY_HOLD_HOLD)

	me := &DlgEdit{wnd, inputs,
		chkPic, wndPic, lblPic,
		lstFrames, btnUncheckAll, btnCheckFilled,
		tags, co.ID_CANCEL}
	defer me.wndPic.HBmp.DeleteObject()

	me.events()
	me.wnd.ShowModal() // blocks until the modal is closed
	return me.result
}

// Relationship between fields and frame IDs.
var FIELD_NAMES = map[uint16]string{
	CHK_ARTIST:      "TPE1",
	CHK_TITLE:       "TIT2",
	CHK_SUBTITLE:    "TIT3",
	CHK_ALBUM:       "TALB",
	CHK_TRACK:       "TRCK",
	CHK_YEAR:        "TYER",
	CHK_GENRE:       "TCON",
	CHK_PERFORMER:   "TPE3",
	CHK_PUBLISHER:   "TPUB",
	CHK_ORIG_ARTIST: "TOPE",
	CHK_ORIG_ALBUM:  "TOAL",
	CHK_ORIG_YEAR:   "TORY",
	CHK_COMPOSER:    "TCOM",
	CHK_LYRICIST:    "TEXT",
	CHK_COMMENT:     "COMM",
}
