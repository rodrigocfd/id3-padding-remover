package main

import (
	"fmt"
	"id3fit/dlgrun"
	"id3fit/id3v2"
)

// Tag operations to be performed.
type TAG_OP int

const (
	TAG_OP_LOAD TAG_OP = iota
	TAG_OP_SAVE_AND_RELOAD
)

// Error report for tag operations.
type TagOpError struct {
	mp3 string
	err error
}

// Opens the DlgRun modal window to perform the chosen operation.
func (me *DlgMain) modalTagOp(mp3s []string, ops TAG_OP) *TagOpError {
	loadOp := func(mp3s []string, cachedTags map[string]*id3v2.Tag) *TagOpError {
		loadedTags := make([]*id3v2.Tag, 0, len(mp3s))

		for _, mp3 := range mp3s {
			if tag, err := id3v2.TagReadFromFile(mp3); err != nil {
				return &TagOpError{ // no further files will be parsed
					mp3: mp3,
					err: fmt.Errorf("load fail: %w", err),
				}
			} else {
				loadedTags = append(loadedTags, tag)
			}
		}

		for i := range mp3s { // atomically cache (or re-cache) the loaded tags
			cachedTags[mp3s[i]] = loadedTags[i]
		}

		return nil
	}

	saveOp := func(mp3s []string, cachedTags map[string]*id3v2.Tag) *TagOpError {
		for _, mp3 := range mp3s {
			tag := cachedTags[mp3]
			if err := tag.SerializeToFile(mp3); err != nil {
				return &TagOpError{ // no further files will be saved
					mp3: mp3,
					err: fmt.Errorf("save fail: %w", err),
				}
			}
		}

		return nil
	}

	var tagOpErr *TagOpError

	dlgrun.NewDlgRun().
		Show(me.wnd, func() {
			switch ops {
			case TAG_OP_LOAD:
				tagOpErr = loadOp(mp3s, me.cachedTags)
			case TAG_OP_SAVE_AND_RELOAD:
				if tagOpErr = saveOp(mp3s, me.cachedTags); tagOpErr != nil {
					tagOpErr = loadOp(mp3s, me.cachedTags)
				}
			}
		})

	return tagOpErr
}
