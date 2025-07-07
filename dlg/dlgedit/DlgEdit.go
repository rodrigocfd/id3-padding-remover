//go:build windows

package dlgedit

import (
	"id3fit/dlg/ids"
	"id3fit/dlg/wndpicture"
	"id3fit/id3v2"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/co"
)

// Modal dialog to edit the fields of one or many tags.
type DlgEdit struct {
	wnd            *ui.Modal
	inputs         []CheckInput // checkbox + input fields
	chkPick        *ui.CheckBox
	wndPic         *wndpicture.WndPicture
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
func ShowNew(parent ui.Parent, tags []*id3v2.Tag) co.ID {
	wnd := ui.NewModalDlg(parent, ids.DLG_EDIT)

	inputs := make([]CheckInput, 0, (ids.TXT_COMMENT-ids.CHK_ARTIST+1)/2)
	for id := ids.CHK_ARTIST; id <= ids.TXT_COMMENT; id += 2 {
		chk := ui.NewCheckBoxDlg(wnd, id, ui.LAY_NONE_NONE)
		var inp ui.ChildControl
		if id == ids.CHK_GENRE {
			inp = ui.NewComboBoxDlg(wnd, ids.CMB_GENRE, ui.LAY_NONE_NONE)
		} else {
			inp = ui.NewEditDlg(wnd, id+1, ui.LAY_NONE_NONE)
		}
		inputs = append(inputs, CheckInput{chk, inp})
	}

	chkPic := ui.NewCheckBoxDlg(wnd, ids.CHK_PICTURE, ui.LAY_NONE_NONE)
	wndPic := wndpicture.New(wnd, ui.DpiX(420), ui.DpiY(60), ui.DpiX(200), ui.DpiY(200))
	lblPic := ui.NewStaticDlg(wnd, ids.LBL_IMAGE_DESCR, ui.LAY_NONE_NONE)

	lstFrames := ui.NewListViewDlg(wnd, ids.LST_FRAMES, ids.MNU_FRAMES, ui.LAY_NONE_NONE)
	btnUncheckAll := ui.NewButtonDlg(wnd, ids.BTN_UNCHECK_ALL, ui.LAY_NONE_NONE)
	btnCheckFilled := ui.NewButtonDlg(wnd, ids.BTN_CHECK_FILLED, ui.LAY_NONE_NONE)

	me := &DlgEdit{wnd, inputs,
		chkPic, wndPic, lblPic,
		lstFrames, btnUncheckAll, btnCheckFilled,
		tags, co.ID_CANCEL}

	me.events()
	me.wnd.ShowModal() // blocks until the modal is closed
	return me.result
}

// Relationship between fields and frame IDs.
var FIELD_NAMES = map[uint16]string{
	ids.CHK_ARTIST:      "TPE1",
	ids.CHK_TITLE:       "TIT2",
	ids.CHK_SUBTITLE:    "TIT3",
	ids.CHK_ALBUM:       "TALB",
	ids.CHK_TRACK:       "TRCK",
	ids.CHK_YEAR:        "TYER",
	ids.CHK_GENRE:       "TCON",
	ids.CHK_PERFORMER:   "TPE3",
	ids.CHK_PUBLISHER:   "TPUB",
	ids.CHK_ORIG_ARTIST: "TOPE",
	ids.CHK_ORIG_ALBUM:  "TOAL",
	ids.CHK_ORIG_YEAR:   "TORY",
	ids.CHK_COMPOSER:    "TCOM",
	ids.CHK_LYRICIST:    "TEXT",
	ids.CHK_COMMENT:     "COMM",
}
