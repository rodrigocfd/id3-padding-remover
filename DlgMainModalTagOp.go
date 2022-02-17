package main

import (
	"fmt"
	"id3fit/dlgrun"
	"id3fit/id3v2"
	"sync"
)

// Tag operations to be performed.
type TAG_OP int

const (
	TAG_OP_LOAD TAG_OP = iota
	TAG_OP_SAVE_AND_RELOAD
)

// Opens the DlgRun modal window to perform the chosen operation.
func (me *DlgMain) modalTagOp(mp3s []string, tagOp TAG_OP) bool {

	loadOp := func(mp3s []string, cachedTags map[string]*id3v2.Tag) []error {
		var waitGroup sync.WaitGroup
		var mutex sync.Mutex
		parsingErrors := make([]error, 0, len(mp3s))
		loadedTags := make(map[string]*id3v2.Tag, len(mp3s))

		for _, mp3 := range mp3s {
			waitGroup.Add(1)
			go func(mp3 string) {
				defer waitGroup.Done()
				if tag, err := id3v2.TagParseFromFile(mp3); err != nil {
					mutex.Lock()
					parsingErrors = append(parsingErrors,
						fmt.Errorf("parsing \"%s\" failed: %w", mp3, err))
					mutex.Unlock()
				} else {
					mutex.Lock()
					loadedTags[mp3] = tag
					mutex.Unlock()
				}
			}(mp3)
		}
		waitGroup.Wait()

		if len(parsingErrors) == 0 { // no errors occurred?
			for mp3, tag := range loadedTags {
				cachedTags[mp3] = tag // atomically cache (or re-cache) the loaded tags
			}
		}
		return parsingErrors
	}

	saveOp := func(mp3s []string, cachedTags map[string]*id3v2.Tag) []error {
		var waitGroup sync.WaitGroup
		var mutex sync.Mutex
		savingErrors := make([]error, 0, len(mp3s))

		for _, mp3 := range mp3s {
			waitGroup.Add(1)
			go func(mp3 string, tag *id3v2.Tag) {
				defer waitGroup.Done()
				if err := tag.SerializeToFile(mp3); err != nil {
					mutex.Lock()
					savingErrors = append(savingErrors,
						fmt.Errorf("saving \"%s\" failed: %w", mp3, err))
					mutex.Unlock()
				}
			}(mp3, cachedTags[mp3])
		}
		waitGroup.Wait()

		return savingErrors
	}

	return dlgrun.NewDlgRun().
		Show(me.wnd, func() []error {
			switch tagOp {
			case TAG_OP_LOAD:
				return loadOp(mp3s, me.cachedTags)
			case TAG_OP_SAVE_AND_RELOAD:
				if errors := saveOp(mp3s, me.cachedTags); len(errors) > 0 {
					return errors
				} else {
					return loadOp(mp3s, me.cachedTags)
				}
			default:
				panic("Invalid TAG_OP.")
			}
		})
}
