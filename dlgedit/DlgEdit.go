//go:build windows

package dlgedit

import (
	"id3fit/dlgpicture"
	"id3fit/id3v2"
	"id3fit/ids"
	"slices"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/ole"
)

// Modal dialog to edit the fields of a tag.
type DlgEdit struct {
	wnd           *ui.Modal
	comRel        ole.Releaser
	pic           ole.IPicture
	tags          []*id3v2.Tag
	orderedFrames []*id3v2.Frame // shallow copy from the tag; only if editing 1 tag

	inputs         []CheckInput // CheckBox + Edit pairs
	lstFrames      *ui.ListView
	btnUncheckAll  *ui.Button
	btnCheckFilled *ui.Button
	dlgPic         *dlgpicture.DlgPicture
	lblPic         *ui.Static

	result co.ID // returned when the user closes the modal
}

type CheckInput struct {
	chk *ui.CheckBox
	in  ui.ChildControl // Edit or ComboBox
}

// Constructor.
func New(parent ui.Parent, tags []*id3v2.Tag) *DlgEdit {
	me := &DlgEdit{
		wnd:    ui.NewModalDlg(parent, ids.DLG_EDIT, false),
		comRel: ole.NewReleaser(), // released after modal closes
		tags:   tags,
	}
	me.loadComPicture()
	if len(tags) == 1 {
		me.orderedFrames = slices.Clone(tags[0].Frames())
	}

	me.inputs = make([]CheckInput, 0, (ids.TXT_COMMENT-ids.CHK_ARTIST+1)/2)
	for id := ids.CHK_ARTIST; id <= ids.TXT_COMMENT; id += 2 {
		chk := ui.NewCheckBoxDlg(me.wnd, id, ui.LAY_NONE_NONE)
		var in ui.ChildControl
		if id == ids.CHK_GENRE {
			in = ui.NewComboBoxDlg(me.wnd, ids.CMB_GENRE, ui.LAY_NONE_NONE)
		} else {
			in = ui.NewEditDlg(me.wnd, id+1, ui.LAY_NONE_NONE)
		}
		me.inputs = append(me.inputs, CheckInput{chk, in})
	}

	me.lstFrames = ui.NewListViewDlg(me.wnd, ids.LST_FRAMES, ids.MNU_FRAMES, ui.LAY_NONE_NONE)
	me.btnUncheckAll = ui.NewButtonDlg(me.wnd, ids.BTN_UNCHECK_ALL, ui.LAY_NONE_NONE)
	me.btnCheckFilled = ui.NewButtonDlg(me.wnd, ids.BTN_CHECK_FILLED, ui.LAY_NONE_NONE)
	me.dlgPic = dlgpicture.New(me.wnd, ui.DpiX(420), ui.DpiY(60), ui.DpiX(200), ui.DpiY(200), &me.pic)
	me.lblPic = ui.NewStaticDlg(me.wnd, ids.LBL_IMAGE_DESCR, ui.LAY_NONE_NONE)
	me.result = co.ID_CANCEL

	me.events()
	return me
}

func (me *DlgEdit) ShowModal() co.ID {
	defer me.comRel.Release()
	me.wnd.ShowModal()
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
