//go:build windows

package dlg

import (
	"fmt"
	"id3fit/id3v2"
	"strings"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/wstr"
	"github.com/rodrigocfd/xslices"
)

func (me *DlgMain) updateTitlebarCount() {
	nFiles := me.lstFiles.ItemCount()
	nSel := me.lstFiles.SelectedItemCount()
	me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)", nSel, nFiles))
}

func (me *DlgMain) addMp3sToList(incomingPaths []string) {
	tagsAndPaths := ShowDlgProgressLoad(me.wnd, incomingPaths)
	if len(tagsAndPaths) > 0 {
		for _, tag := range tagsAndPaths {
			var item ui.ListViewItem
			if existingItem, ok := me.lstFiles.FindItem(tag.Path); ok { // file already loaded?
				item = existingItem // current tag object will be replaced
			} else {
				item = me.lstFiles.AddItem(tag.Path).
					SetIcon16(ui.IcoExt("mp3")) // insert new item
			}
			item.SetData(tag.Tag) // store tag in item
			me.renderMp3InList(item)
		}
		me.sortList()
		me.lstFiles.Col(0).SetWidthToFill()
	}
	me.updateTitlebarCount()
}

// Retrieves the tag from the list view item, and puts its values in the
// columns.
func (me *DlgMain) renderMp3InList(item ui.ListViewItem) {
	pTag := item.Data().(*id3v2.Tag) // retrieve tag stored in item
	if pTag.IsEmpty() {
		item.SetText(1, "N/A") // MP3 without tag
	} else {
		item.SetText(1, fmt.Sprintf("%d", pTag.Padding()))
	}

	if pTag.FrameByName4("APIC") != nil {
		item.SetText(2, "\u2713") // checkmark symbol
	} else {
		item.SetText(2, "")
	}

	item.SetText(3, pTag.ReplayGainStatus())

	me.renderMp3TextCell(item, 4, pTag, "TPE1")
	me.renderMp3TextCell(item, 5, pTag, "TYER")
	me.renderMp3TextCell(item, 6, pTag, "TALB")
	me.renderMp3TextCell(item, 7, pTag, "TRCK")
	me.renderMp3TextCell(item, 8, pTag, "TIT2")
	me.renderMp3TextCell(item, 9, pTag, "TCON")
	me.renderMp3TextCell(item, 10, pTag, "TPE3")
	me.renderMp3TextCell(item, 11, pTag, "TCOM")
	me.renderMp3TextCell(item, 12, pTag, "TEXT")
	me.renderMp3TextCell(item, 13, pTag, "TOPE")

	if pFrame := pTag.FrameByName4("COMM"); pFrame != nil {
		body, _ := pFrame.Body().(*id3v2.BodyComment)
		item.SetText(14, body.Text)
	} else {
		item.SetText(14, "") // clear
	}
}

func (me *DlgMain) renderMp3TextCell(item ui.ListViewItem, colIndex int, tag *id3v2.Tag, name4 string) {
	if pFrame := tag.FrameByName4(name4); pFrame != nil { // such name4 frame exists
		body, _ := pFrame.Body().(*id3v2.BodyText)
		item.SetText(colIndex, body.Text)
	} else {
		item.SetText(colIndex, "") // clear
	}
}

// Sorts the files in the list view according to the current sortCol and sortAsc
// values.
func (me *DlgMain) sortList() {
	me.lstFiles.SortItems(func(itemA, itemB ui.ListViewItem) int {
		cmp := 0
		if me.sortCol == 1 { // by padding size
			tagA := itemA.Data().(*id3v2.Tag)
			tagB := itemB.Data().(*id3v2.Tag)
			cmp = tagA.Padding() - tagB.Padding()
		} else { // by column text
			cmp = wstr.CmpI(itemA.Text(me.sortCol), itemB.Text(me.sortCol))
		}

		if me.sortAsc {
			return cmp
		} else {
			return -cmp
		}
	})
}

// Asks user confirmation to delete APIC and ReplayGain frames. If yes, deletes
// the frames from all selected tags, without saving them to the MP3 files, and
// returns true.
func (me *DlgMain) removePicRg(delRg bool) bool {
	nFiles := me.lstFiles.SelectedItemCount()
	text := fmt.Sprintf("Remove picture frame from %d file(s)?", nFiles)
	if delRg {
		text = fmt.Sprintf("Remove picture and ReplayGain frames from %d file(s)?", nFiles)
	}
	if !ui.MsgOkCancel(me.wnd, "Remove frames", "", text, "&Remove") {
		return false
	}

	for _, item := range me.lstFiles.SelectedItems() {
		pTag := item.Data().(*id3v2.Tag)
		pTag.RemoveFrameIf(func(pFrame *id3v2.Frame) bool {
			if pFrame.Name4() == "APIC" {
				return true
			}
			if delRg && pFrame.Name4() == "TXXX" {
				if pBody, ok := pFrame.Body().(*id3v2.BodyUserText); ok {
					if strings.HasPrefix(pBody.Descr, "replaygain_track_") ||
						strings.HasPrefix(pBody.Descr, "replaygain_album_") {
						return true
					}
				}
			}
			return false
		})
	}
	return true
}

// Displays the edit modal dialog with the selected tags, and returns true of
// the user clicks OK.
func (me *DlgMain) editSelected() bool {
	if me.lstFiles.SelectedItemCount() == 0 {
		return false // Enter key will hit here even without selected items
	}

	clonedTags := xslices.Map(me.lstFiles.SelectedItems(), func(_ int, item ui.ListViewItem) *id3v2.Tag {
		pTag := item.Data().(*id3v2.Tag)
		return pTag.Clone()
	})

	if ShowDlgEdit(me.wnd, clonedTags) == co.ID_OK {
		for i, item := range me.lstFiles.SelectedItems() {
			item.SetData(clonedTags[i]) // replace the selected tags with the cloned, edited ones
		}
		return true
	}

	return false
}

// Displays the save modal dialog, and saves the selected tags to the MP3 files.
func (me *DlgMain) saveSelected() {
	selTags := xslices.Map(
		me.lstFiles.SelectedItems(),
		func(_ int, item ui.ListViewItem) TagAndPath {
			return TagAndPath{
				Tag:  item.Data().(*id3v2.Tag),
				Path: item.Text(0),
			}
		},
	)
	ShowDlgProgressSave(me.wnd, selTags)

	for _, item := range me.lstFiles.SelectedItems() {
		me.renderMp3InList(item) // re-render, all paddings have been removed
	}
	me.sortList()
}
