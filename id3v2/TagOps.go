//go:build windows

package id3v2

import (
	"fmt"
	"runtime"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/xslices"
)

type TagAndPath struct {
	Tag  *Tag
	Path string
}

// Loads the tags from multiple paths, which can be MP3 files or folders.
func LoadMultipleTagsAsync(
	paths []string,
	onDone func(tagsAndPaths []TagAndPath, err error),
) {
	go func() {
		runtime.LockOSThread()
		mp3Paths := make([]string, 0, len(paths)) // arbitrary
		for _, path := range paths {
			if win.PathIsFolder(path) {
				err := win.PathEnumDeepFunc(path, func(fileInfo *win.PathEnumInfo) error {
					if win.PathHasExtension(fileInfo.Path, "mp3") {
						mp3Paths = append(mp3Paths, fileInfo.Path)
					}
					return nil
				})
				if err != nil {
					onDone(nil, fmt.Errorf("file enumeration failed: %w", err))
					return
				}
			} else if win.PathHasExtension(path, "mp3") {
				mp3Paths = append(mp3Paths, path)
			}
		}

		if len(mp3Paths) == 0 {
			onDone([]TagAndPath{}, nil) // no MP3 files found
			return
		}

		tagsAndPaths := make([]TagAndPath, 0, len(mp3Paths))
		for _, mp3Path := range mp3Paths {
			pTag, err := TagFromFile(mp3Path)
			if err != nil {
				onDone(nil, fmt.Errorf("failed to load tag from: %s: %w", mp3Path, err))
				return
			}
			tagsAndPaths = append(tagsAndPaths, TagAndPath{pTag, mp3Path})
		}
		onDone(tagsAndPaths, nil) // all files loaded successfully
	}()
}

// Saves the tags to the MP3 files.
func SaveMultipleTagsAsync(
	tagsAndPaths []TagAndPath,
	onDone func(errors []error),
) {
	go func() {
		runtime.LockOSThread()
		errors := []error{}
		for _, tagAndPath := range tagsAndPaths {
			if err := tagAndPath.Tag.SaveToFile(tagAndPath.Path); err != nil {
				errors = append(errors,
					fmt.Errorf("failed to save %s: %w", tagAndPath.Path, err)) // save error and keep going
			}
		}
		onDone(errors)
	}()
}

// If the frame is the same across all tags, returns it; otherwise returns nil.
func SameFrameAcrossAllTags(name4 string, tags []*Tag) *Frame {
	if len(tags) == 0 {
		return nil // no tags, no frame
	} else if len(tags) == 1 {
		return tags[0].FrameByName4(name4) // only 1 tag, direct query
	}

	pFrame0 := tags[0].FrameByName4(name4) // query the first tag right away
	if pFrame0 == nil {
		return nil // first tag doesn't have the frame, stop right now
	}

	allSame := xslices.EveryFunc(tags[1:], func(_ int, pTag *Tag) bool { // skip first tag
		pFrame := pTag.FrameByName4(name4)
		if pFrame == nil {
			return false // frame doesn't exist in this posterior tag
		} else {
			// Compare the textual rendering of this frame with the first frame.
			return pFrame.Body().AsText() == pFrame0.Body().AsText()
		}
	})

	if allSame {
		return pFrame0
	}
	return nil
}
