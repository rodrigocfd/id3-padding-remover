package dlgmain

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
//
// Returns false if any error occurred. Error messages are displayed in the
// DlgRun itself.
func (me *DlgMain) modalTagOp(tagOp TAG_OP, mp3s ...string) bool {
	dlg := dlgrun.NewDlgRun()
	return dlg.Show(me.wnd, func() []error {
		switch tagOp {
		case TAG_OP_LOAD:
			return loadOp(me.cachedTags, mp3s...)
		case TAG_OP_SAVE_AND_RELOAD:
			if errors := saveOp(me.cachedTags, mp3s...); len(errors) > 0 {
				return errors
			} else {
				return loadOp(me.cachedTags, mp3s...)
			}
		default:
			panic("Invalid TAG_OP.")
		}
	})
}

// Loads the tags from the MP3 files.
func loadOp(cachedTags map[string]*id3v2.Tag, mp3s ...string) []error {
	var waitGroup sync.WaitGroup
	var mutex sync.Mutex
	parsingErrors := make([]error, 0, len(mp3s))
	loadedTags := make(map[string]*id3v2.Tag, len(mp3s))

	for _, mp3 := range mp3s {
		waitGroup.Add(1)
		go func(mp3 string) { // spawn one goroutine per file
			defer waitGroup.Done()
			if tag, err := id3v2.NewTagParseFromFile(mp3); err != nil {
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

// Saves the tags into the MP3 files.
func saveOp(cachedTags map[string]*id3v2.Tag, mp3s ...string) []error {
	savingErrors := make([]error, 0, len(mp3s))
	for _, mp3 := range mp3s { // save sequentially to stay safe from writing errors
		tag := cachedTags[mp3]
		if err := tag.SerializeToFile(mp3); err != nil {
			savingErrors = append(savingErrors,
				fmt.Errorf("saving \"%s\" failed: %w", mp3, err)) // save error and keep going
		}
	}
	return savingErrors
}