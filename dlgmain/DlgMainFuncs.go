//go:build windows

package dlgmain

import (
	"fmt"
	"id3fit/dlgedit"
	"id3fit/id3v2"
	"slices"
	"strings"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
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
		if win.Path.IsFolder(incomingPath) {
			allPaths = append(allPaths,
				slices.Collect(win.Path.IterFilesDeep(incomingPath))...)
		} else {
			allPaths = append(allPaths, incomingPath)
		}
	}

	nonMp3Count := 0 // count how many non-MP3 we have
	for _, path := range allPaths {
		if !win.Path.HasExtension(path, "mp3") {
			nonMp3Count++
		}
	}
	if nonMp3Count == len(allPaths) { // zero MP3s found?
		me.wnd.Hwnd().MessageBox(
			fmt.Sprintf("No MP3 found amongst %d files.", len(allPaths)),
			"No MP3 files", co.MB_ICONERROR)
		return // nothing do to
	}

	tags := make([]*id3v2.Tag, 0, len(allPaths)) // cache all the tags
	for _, path := range allPaths {
		if win.Path.HasExtension(path, "mp3") { // ignore non-MP3 files
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
		me.tags[item.Uid()] = tag // store tag in cache
		me.renderMp3InList(item)
	}
	me.sortList()
	me.lstFiles.Cols.Get(0).SetWidthToFill()
}

func (me *DlgMain) renderMp3InList(item ui.ListViewItem) {
	tag := me.tags[item.Uid()] // retrieve tag from cache

	if tag.IsEmpty() {
		item.SetText(1, "N/A") // MP3 without tag
	} else {
		item.SetText(1, fmt.Sprintf("%d", tag.Padding()))
	}

	if tag.FrameByName4("APIC") != nil {
		item.SetText(2, "\u2713") // checkmark
	} else {
		item.SetText(2, "")
	}

	item.SetText(3, tag.ReplayGainStatus())

	me.renderMp3TextColumn(item, 4, tag, "TPE1")
	me.renderMp3TextColumn(item, 5, tag, "TYER")
	me.renderMp3TextColumn(item, 6, tag, "TALB")
	me.renderMp3TextColumn(item, 7, tag, "TRCK")
	me.renderMp3TextColumn(item, 8, tag, "TIT2")
	me.renderMp3TextColumn(item, 9, tag, "TCON")
	me.renderMp3TextColumn(item, 10, tag, "TPE3")
	me.renderMp3TextColumn(item, 11, tag, "TCOM")
	me.renderMp3TextColumn(item, 12, tag, "TEXT")
	me.renderMp3TextColumn(item, 13, tag, "TOPE")

	if frame := tag.FrameByName4("COMM"); frame != nil {
		body, _ := frame.Body().(*id3v2.BodyComment)
		item.SetText(14, body.Text)
	} else {
		item.SetText(14, "") // clear
	}
}

func (me *DlgMain) renderMp3TextColumn(item ui.ListViewItem, colIndex int, tag *id3v2.Tag, name4 string) {
	if frame := tag.FrameByName4(name4); frame != nil {
		body, _ := frame.Body().(*id3v2.BodyText)
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
			tagA := me.tags[itemA.Uid()]
			tagB := me.tags[itemB.Uid()]
			cmp = int(tagA.Padding()) - int(tagB.Padding())
		} else { // by column text
			cmp = win.Str.CmpI(itemA.Text(me.sortCol), itemB.Text(me.sortCol))
		}

		if me.sortAsc {
			return cmp
		} else {
			return -cmp
		}
	})
}

func (me *DlgMain) removePicRg(delRg bool) {
	nFiles := me.lstFiles.Items.SelectedCount()
	msg := fmt.Sprintf("Remove picture frame from %d file(s)?", nFiles)
	if delRg {
		msg = fmt.Sprintf("Remove picture and ReplayGain frames from %d file(s)?", nFiles)
	}

	ret, _ := win.TaskDialogIndirect(win.TASKDIALOGCONFIG{
		HwndParent:  me.wnd.Hwnd(),
		WindowTitle: "Remove frames",
		Content:     msg,
		HMainIcon:   win.TdcIconTdi(co.TDICON_WARNING),
		Flags:       co.TDF_ALLOW_DIALOG_CANCELLATION | co.TDF_POSITION_RELATIVE_TO_WINDOW,
		Buttons: []win.TASKDIALOG_BUTTON{
			{Id: co.ID_OK, Text: "&Remove"},
			{Id: co.ID_CANCEL, Text: "&Cancel"},
		},
	})
	if ret != co.ID_OK {
		return
	}

	for item := range me.lstFiles.Items.IterSelected() {
		tag := me.tags[item.Uid()]
		tag.RemoveFrameIf(func(frame *id3v2.Frame) bool {
			if frame.Name4() == "APIC" {
				return true
			}
			if delRg && frame.Name4() == "TXXX" {
				if body, ok := frame.Body().(*id3v2.BodyUserText); ok {
					if strings.HasPrefix(body.Descr, "replaygain_track_") ||
						strings.HasPrefix(body.Descr, "replaygain_album_") {
						return true
					}
				}
			}
			return false
		})
	}
}

func (me *DlgMain) editSelected() co.ID {
	if me.lstFiles.Items.SelectedCount() == 0 {
		return co.ID_CANCEL // Enter key will hit here even without selected items
	}

	selTags := make([]*id3v2.Tag, 0, me.lstFiles.Items.SelectedCount())
	for item := range me.lstFiles.Items.IterSelected() {
		selTags = append(selTags, me.tags[item.Uid()])
	}
	wndEdit := dlgedit.New(me.wnd, selTags)
	return wndEdit.ShowModal()
}

func (me *DlgMain) saveSelected() {
	type Fail struct {
		file string
		err  error
	}
	failed := make([]Fail, 0)

	for item := range me.lstFiles.Items.IterSelected() {
		tag := me.tags[item.Uid()]
		if err := tag.SaveToFile(); err != nil {
			failed = append(failed, Fail{file: tag.Path(), err: err})
		}
		me.renderMp3InList(item)
	}
	me.sortList()

	if len(failed) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%d file(s) failed to save:", len(failed)))
		for _, fail := range failed {
			sb.WriteString("\n\n")
			sb.WriteString(fail.file)
			sb.WriteString("\n")
			sb.WriteString(fail.err.Error())
		}
		win.TaskDialogIndirect(win.TASKDIALOGCONFIG{
			HwndParent:    me.wnd.Hwnd(),
			WindowTitle:   "Error saving file(s)",
			Content:       sb.String(),
			HMainIcon:     win.TdcIconTdi(co.TDICON_ERROR),
			CommonButtons: co.TDCBF_OK,
			Flags:         co.TDF_ALLOW_DIALOG_CANCELLATION | co.TDF_POSITION_RELATIVE_TO_WINDOW,
		})
	}
}
