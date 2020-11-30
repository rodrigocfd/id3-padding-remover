package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"windigo/ui"
)

func parseTagHeader(src []byte) (int, error) {
	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return 0, errors.New("No ID3 tag found.")
	}

	// Validate tag version 2.3.0.
	if !bytes.Equal(src[3:5], []byte{3, 0}) { // the first "2" is not stored in the tag
		return 0, errors.New(
			fmt.Sprintf("Tag version 2.%d.%d is not supported, only 2.3.0.",
				src[3], src[4]),
		)
	}

	// Validate unsupported flags.
	if (src[5] & 0b1000_0000) != 0 {
		return 0, errors.New("Tag is unsynchronised, not supported.")
	} else if (src[5] & 0b0100_0000) != 0 {
		return 0, errors.New("Tag extended header not supported.")
	}

	// Read tag size.
	totalTagSize := int(_Util.SynchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]), // also count 10-byte tag header
	) + 10)

	return totalTagSize, nil
}

func parseAllFrames(src []byte) ([]Frame, int, error) {
	frames := make([]Frame, 0, 6) // arbitrary capacity
	paddingSize := 0

	for {
		if len(src) == 0 { // end of tag, no padding found
			break
		} else if _Util.IsSliceZeroed(src) { // we entered a padding region after all frames
			paddingSize = len(src) // store padding size
			break
		}

		newFrame, totalFrameSize, err := _ParseFrame(src)
		if err != nil {
			return nil, 0, err // error when parsing the frame
		}
		frames = append(frames, newFrame) // add the frame to our collection

		if totalFrameSize > len(src) { // means the tag was serialized with error
			return nil, 0, errors.New("Frame size is greater than real size.")
		}

		src = src[totalFrameSize:] // now starts at 1st byte of next frame
	}

	return frames, paddingSize, nil // all frames parsed successfully
}

func (me *Tag) writeTagToFile(mp3Path string, newTagBlob []byte) error {
	fout, err := ui.OpenFileMapped(mp3Path, ui.FILEMAP_MODE_RW)
	if err != nil {
		return err
	}
	defer fout.Close()
	fileMem := fout.HotSlice()

	currentTag, err := ParseTagFromBinary(fileMem)
	if err != nil {
		return err
	}

	diff := len(newTagBlob) - currentTag.TotalTagSize() // size difference between new/old tags

	if diff > 0 { // new tag is larger, we need to make room
		if err := fout.SetSize(fout.Size() + diff); err != nil {
			return err
		}
	}

	// Move the MP3 data block inside the file.
	copy(fileMem[int(currentTag.TotalTagSize())+diff:], fileMem[currentTag.TotalTagSize():])

	// Copy the new tag into the file, no padding.
	copy(fileMem, newTagBlob)

	if diff < 0 { // new tag is shorter, shrink
		if err := fout.SetSize(fout.Size() + diff); err != nil {
			return err
		}
	}

	return nil
}
