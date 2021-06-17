package id3

import (
	"errors"
	"fmt"
	"id3fit/id3/util"
)

// Frame is polymorphic: the underlying type will expose the methods to access
// the contents.
type Frame interface {
	Name4() string
	OriginalSize() int
	Serialize() []byte
}

// Constructor.
func _ParseFrame(src []byte) (Frame, error) {
	frameBase := _ParseFrameBase(src)
	src = src[10:frameBase.OriginalSize()] // skip frame header, truncate to frame size

	if frameBase.Name4() == "COMM" {
		return _ParseFrameComment(frameBase, src)

	} else if frameBase.Name4()[0] == 'T' {
		texts, e := util.ParseAnyStrings(src)
		if e != nil {
			return nil, e
		}

		if len(texts) == 0 {
			return nil, errors.New(
				fmt.Sprintf("Frame %s contains no texts.", frameBase.Name4()))

		} else if len(texts) == 1 {
			return _ParseFrameText(frameBase, texts)

		} else {
			return _ParseFrameMultiText(frameBase, texts)
		}
	}

	// Anything else is treated as raw binary.
	return _ParseFrameBinary(frameBase, src)
}
