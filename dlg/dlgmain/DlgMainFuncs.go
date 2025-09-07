//go:build windows

package dlgmain

import (
	"fmt"
	"id3fit/dlg/dlgedit"
	"id3fit/id3v2"
	"strings"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/wstr"
	"github.com/rodrigocfd/xslices"
)

func (me *DlgMain) setWaitState(set bool) {
	if set {
		me.wnd.Hwnd().SetWindowText("Working...")
		me.lstFiles.Hwnd().EnableWindow(false)
		me.isWaiting = true
	} else {
		me.updateTitlebarCount()
		me.lstFiles.Hwnd().EnableWindow(true)
		me.isWaiting = false // read in WM_SETCURSOR
		cPos, _ := win.GetCursorPos()
		win.SetCursorPos(int(cPos.X), int(cPos.Y)) // force cursor redraw
	}
}

func (me *DlgMain) updateTitlebarCount() {
	nFiles := me.lstFiles.Items.Count()
	nSel := me.lstFiles.Items.SelectedCount()
	me.wnd.Hwnd().SetWindowText(fmt.Sprintf("ID3 Fit (%d/%d)", nSel, nFiles))
}

func (me *DlgMain) addMp3sToListAsync(incomingPaths []string) {
	me.setWaitState(true)
	go func() {
		allPaths := make([]string, 0, len(incomingPaths)) // grab all files within all subfolders
		for _, incomingPath := range incomingPaths {
			if win.PathIsFolder(incomingPath) {
				nested, _ := win.EnumFilesDeep(incomingPath)
				allPaths = append(allPaths, nested...)
			} else {
				allPaths = append(allPaths, incomingPath)
			}
		}

		nonMp3Count := xslices.CountFunc(allPaths, func(_ int, path string) bool { // count how many non-MP3 we have
			return !win.PathHasExtension(path, "mp3")
		})
		if nonMp3Count == len(allPaths) { // zero MP3s found?
			me.wnd.UiThread(func() {
				me.wnd.Hwnd().MessageBox(
					fmt.Sprintf("No MP3 found amongst %d files.", len(allPaths)),
					"No MP3 files", co.MB_ICONERROR)
				me.setWaitState(false)
			})
			return // nothing do to
		}

		type TagAndPath struct { // a tag and its file path
			pTag *id3v2.Tag
			path string
		}

		mp3ToReadCount := len(allPaths) - nonMp3Count
		tags := make([]TagAndPath, 0, mp3ToReadCount) // load and cache all the MP3 tags
		for idxMp3, path := range allPaths {
			if win.PathHasExtension(path, "mp3") { // ignore non-MP3 files
				pTag, err := id3v2.TagFromFile(path)
				if err != nil {
					me.wnd.UiThread(func() {
						me.wnd.Hwnd().MessageBox(
							fmt.Sprintf("Error loading tag:\n%s\n\n%s", path, err.Error()),
							"Error", co.MB_ICONERROR)
						me.setWaitState(false)
					})
					return // stop on first error, no tag is loaded
				}
				tags = append(tags, TagAndPath{
					pTag: pTag,
					path: path,
				})

				me.wnd.UiThread(func() { // UI progress feedback
					me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%d/%d files read...", idxMp3+1, mp3ToReadCount))
				})
			}
		}

		me.wnd.UiThread(func() { // finally fill the listview with the tags
			for _, tag := range tags {
				var item ui.ListViewItem
				if existingItem, ok := me.lstFiles.Items.Find(tag.path); ok { // file already loaded?
					item = existingItem // current tag object will be replaced
				} else {
					item = me.lstFiles.Items.AddWithIcon(0, tag.path) // insert new item
				}
				item.SetData(tag.pTag) // store tag in item
				me.renderMp3InList(item)
			}
			me.sortList()
			me.lstFiles.Cols.Get(0).SetWidthToFill()
			me.setWaitState(false)
		})
	}()
}

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

	clonedTags := xslices.Map(me.lstFiles.Items.Selected(), func(_ int, item ui.ListViewItem) *id3v2.Tag {
		pTag := item.Data().(*id3v2.Tag)
		return pTag.Clone()
	})

	if dlgedit.ShowNew(me.wnd, clonedTags) == co.ID_OK {
		for i, item := range me.lstFiles.Items.Selected() {
			item.SetData(clonedTags[i]) // replace the selected tags with the edited ones
		}
		return true
	}

	return false
}

func (me *DlgMain) saveSelectedAsync() {
	type (
		Failure struct { // a register of one saving error
			path string
			err  error
		}
		TagAndPath struct { // a tag and its file path
			pTag *id3v2.Tag
			path string
		}
	)

	failures := make([]Failure, 0) // to store all failures we get

	selTags := xslices.Map(me.lstFiles.Items.Selected(), func(_ int, item ui.ListViewItem) TagAndPath {
		return TagAndPath{
			pTag: item.Data().(*id3v2.Tag),
			path: item.Text(0),
		}
	})
	me.setWaitState(true)

	go func() {
		for idxTag, selTag := range selTags { // process each tag sequentially
			if err := selTag.pTag.SaveToFile(selTag.path); err != nil {
				failures = append(failures, Failure{selTag.path, err}) // store error, and keep going
			}

			me.wnd.UiThread(func() { // UI progress feedback
				me.wnd.Hwnd().SetWindowText(fmt.Sprintf("%d/%d files written...", idxTag+1, len(selTags)))
			})
		}

		me.wnd.UiThread(func() { // UI final feedback
			for _, item := range me.lstFiles.Items.Selected() {
				me.renderMp3InList(item) // re-render, all paddings have been removed
			}
			me.sortList()

			if len(failures) > 0 { // any errors?
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("%d file(s) failed to save:", len(failures)))
				for _, fail := range failures {
					sb.WriteString("\n\n")
					sb.WriteString(fail.path)
					sb.WriteString("\n")
					sb.WriteString(fail.err.Error())
				}
				ui.MsgError(me.wnd, "Error saving file(s)", "", sb.String())
			}
			me.setWaitState(false)
			me.lstFiles.Focus()
		})
	}()
}
