//go:build windows

package dlgmain

import (
	"fmt"
	"id3fit/dlgedit"
	"id3fit/id3v2"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/wstr"
)

func (me *DlgMain) withWaitCursor(fun func()) {
	me.wnd.Hwnd().SetWindowText("Loading...")
	me.wnd.Hwnd().EnableWindow(false)
	hCursorWait, _ := win.HINSTANCE(0).LoadCursor(win.CursorResIdc(co.IDC_WAIT))
	hCursorOrig, _ := hCursorWait.SetCursor()

	fun()

	hCursorOrig.SetCursor()
	me.wnd.Hwnd().EnableWindow(true)
	me.updateTitlebarCount()
}

func (me *DlgMain) addMp3sToList(incomingPaths []string) {
	allPaths := make([]string, 0, len(incomingPaths)) // grab all files within all subfolders
	for _, incomingPath := range incomingPaths {
		if win.PathIsFolder(incomingPath) {
			nested, _ := win.EnumFilesDeep(incomingPath)
			allPaths = append(allPaths, nested...)
		} else {
			allPaths = append(allPaths, incomingPath)
		}
	}

	nonMp3Count := 0 // count how many non-MP3 we have
	for _, path := range allPaths {
		if !win.PathHasExtension(path, "mp3") {
			nonMp3Count++
		}
	}
	if nonMp3Count == len(allPaths) { // zero MP3s found?
		me.wnd.Hwnd().MessageBox(
			fmt.Sprintf("No MP3 found amongst %d files.", len(allPaths)),
			"No MP3 files", co.MB_ICONERROR)
		return // nothing do to
	}

	tags := make([]*id3v2.Tag, 0, len(allPaths)-nonMp3Count) // cache all the tags
	for _, path := range allPaths {
		if win.PathHasExtension(path, "mp3") { // ignore non-MP3 files
			tag, err := id3v2.LoadTag(path)
			if err != nil {
				me.wnd.Hwnd().MessageBox(
					fmt.Sprintf("Error loading tag:\n%s\n\n%s", path, err.Error()),
					"Error", co.MB_ICONERROR)
				return // on error, no tag is loaded
			}
			tags = append(tags, tag)
		}
	}

	for _, tag := range tags {
		var item ui.ListViewItem
		if existingItem, ok := me.lstFiles.Items.Find(tag.Path()); ok {
			item = existingItem // current tag object will be replaced
		} else {
			item = me.lstFiles.Items.AddWithIcon(0, tag.Path()) // insert new item
		}
		item.SetData(tag) // store tag in item
		me.renderMp3InList(item)
	}
	me.sortList()
	me.lstFiles.Cols.Get(0).SetWidthToFill()
}

func (me *DlgMain) renderMp3InList(item ui.ListViewItem) {
	pTag := item.Data().(*id3v2.Tag) // retrieve tag stored in item
	if pTag.IsEmpty() {
		item.SetText(1, "N/A") // MP3 without tag
	} else {
		item.SetText(1, fmt.Sprintf("%d", pTag.Padding()))
	}

	if pTag.FrameByName4("APIC") != nil {
		item.SetText(2, "\u2713") // checkmark
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
	if pFrame := tag.FrameByName4(name4); pFrame != nil {
		body, _ := pFrame.Body().(*id3v2.BodyText)
		item.SetText(colIndex, body.Text)
	} else {
		item.SetText(colIndex, "") // clear
	}
}

func (me *DlgMain) updateTitlebarCount() {
	nFiles := me.lstFiles.Items.Count()
	nSel := me.lstFiles.Items.SelectedCount()
	me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)", nSel, nFiles))
}

func (me *DlgMain) sortList() {
	me.lstFiles.Items.Sort(func(itemA, itemB ui.ListViewItem) int {
		cmp := 0
		if me.sortCol == 1 { // by padding size
			tagA := itemA.Data().(*id3v2.Tag)
			tagB := itemB.Data().(*id3v2.Tag)
			cmp = int(tagA.Padding()) - int(tagB.Padding())
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

func (me *DlgMain) removePicRg(delRg bool) bool {
	nFiles := me.lstFiles.Items.SelectedCount()
	text := fmt.Sprintf("Remove picture frame from %d file(s)?", nFiles)
	if delRg {
		text = fmt.Sprintf("Remove picture and ReplayGain frames from %d file(s)?", nFiles)
	}
	if ui.MsgOkCancel(me.wnd, "Remove frames", "", text, "&Remove") != co.ID_OK {
		return false
	}

	for _, item := range me.lstFiles.Items.Selected() {
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

func (me *DlgMain) editSelected() bool {
	if me.lstFiles.Items.SelectedCount() == 0 {
		return false // Enter key will hit here even without selected items
	}

	clonedTags := make([]*id3v2.Tag, 0, me.lstFiles.Items.SelectedCount())
	for _, item := range me.lstFiles.Items.Selected() {
		pTag := item.Data().(*id3v2.Tag)
		clonedTags = append(clonedTags, pTag.Clone())
	}

	if dlgedit.Show(me.wnd, clonedTags) == co.ID_OK {
		for i, item := range me.lstFiles.Items.Selected() {
			item.SetData(clonedTags[i]) // replace the selected tags with the edited ones
		}
		return true
	}

	return false
}

func (me *DlgMain) saveSelected() {
	type SaveFail struct {
		file string
		err  error
	}
	saveFails := make([]SaveFail, 0)

	for _, item := range me.lstFiles.Items.Selected() {
		pTag := item.Data().(*id3v2.Tag)
		if err := pTag.SaveToFile(); err != nil {
			saveFails = append(saveFails, SaveFail{pTag.Path(), err})
		}
		me.renderMp3InList(item)
	}
	me.sortList()

	if len(saveFails) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%d file(s) failed to save:", len(saveFails)))
		for _, fail := range saveFails {
			sb.WriteString("\n\n")
			sb.WriteString(fail.file)
			sb.WriteString("\n")
			sb.WriteString(fail.err.Error())
		}
		ui.MsgError(me.wnd, "Error saving file(s)", "", sb.String())
	}
}
