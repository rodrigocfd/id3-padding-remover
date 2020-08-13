package id3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type _Parser struct {
	version     [3]uint16
	tagSize     uint32
	paddingSize uint32
	frames      []Frame
}

func (me *_Parser) Version() [3]uint16  { return me.version }
func (me *_Parser) TagSize() uint32     { return me.tagSize }
func (me *_Parser) PaddingSize() uint32 { return me.paddingSize }
func (me *_Parser) Frames() []Frame     { return me.frames }

func (me *_Parser) Parse(src []byte) error {
	if err := me.parseTagHeader(src); err != nil {
		return err
	}

	src = src[10:me.tagSize] // skip 10-byte tag header; truncate to tag size
	if err := me.parseAllFrames(src); err != nil {
		return err
	}

	return nil
}

func (me *_Parser) parseTagHeader(src []byte) error {
	// Check ID3 magic bytes.
	if !bytes.Equal(src[:3], []byte("ID3")) {
		return errors.New("No ID3 tag found.")
	}

	// Validate tag version 2.3.0.
	me.version = [3]uint16{
		2, // the "2" is not actually stored in the tag itself
		uint16(src[3]),
		uint16(src[4]),
	}
	if me.version[1] != 3 && me.version[2] != 0 { // not v2.3.0?
		return errors.New(
			fmt.Sprintf("Tag version 2.%d.%d is not supported, only 2.3.0.",
				me.version[1], me.version[2]),
		)
	}

	// Validade unsupported flags.
	if (src[5] & 0b1000_0000) != 0 {
		return errors.New("Tag is unsynchronised, not supported.")
	} else if (src[5] & 0b0100_0000) != 0 {
		return errors.New("Tag extended header not supported.")
	}

	// Read tag size.
	me.tagSize = _Util.SynchSafeDecode(
		binary.BigEndian.Uint32(src[6:10]), // also count 10-byte tag header
	) + 10

	return nil
}

func (me *_Parser) parseAllFrames(src []byte) error {
	for {
		if len(src) == 0 { // end of tag, no padding found
			break
		} else if _Util.IsSliceZeroed(src) { // we entered a padding region after all frames
			me.paddingSize = uint32(len(src)) // store padding size
			break
		}

		newFrame, err := me.parseFrame(src)
		if err != nil {
			return err // error when parsing the frame
		}
		me.frames = append(me.frames, newFrame)

		if int(newFrame.FrameSize()) > len(src) {
			return errors.New("Frame size is greater than real size.")
		}

		src = src[newFrame.FrameSize():] // now starts at 1st byte of next frame
	}
	return nil
}

func (me *_Parser) parseFrame(src []byte) (Frame, error) {
	baseFr := _BaseFrame{}
	baseFr.name4 = string(src[0:4])
	baseFr.frameSize = binary.BigEndian.Uint32(src[4:8]) + 10 // also count 10-byte tag header

	src = src[10:baseFr.frameSize] // skip frame header, limit to frame size

	var finalFr Frame
	var err error = nil

	if baseFr.name4 == "COMM" { // comment
		finalFr, err = me.parseCommentFrame(src)
		finalFr.(*FrameComment)._BaseFrame = baseFr

	} else if baseFr.name4[0] == 'T' { // single or multi text
		var texts []string
		texts, err = me.parseTextFrame(src) // retrieve all texts

		if len(texts) == 1 { // single text
			finalFr = &FrameText{}
			finalFr.(*FrameText)._BaseFrame = baseFr
			finalFr.(*FrameText).text = texts[0]

		} else { // multi text
			finalFr = &FrameMultiText{}
			finalFr.(*FrameMultiText)._BaseFrame = baseFr
			finalFr.(*FrameMultiText).texts = texts
		}

	} else { // anything else is treated as raw binary
		finalFr = me.parseBinaryFrame(src)
		finalFr.(*FrameBinary)._BaseFrame = baseFr
	}

	if err != nil {
		return nil, err
	}
	return finalFr, nil // frame parsed successfully
}

func (me *_Parser) parseCommentFrame(src []byte) (*FrameComment, error) {
	fr := &FrameComment{}

	// Retrieve text encoding.
	if src[0] != 0x00 && src[0] != 0x01 {
		return nil, errors.New("Unknown comment encoding.")
	}
	isUtf16 := src[0] == 0x01
	src = src[1:] // skip encoding byte

	// Retrieve language string, always ASCII.
	fr.lang = _Util.ConvertAsciiStrings(src[:3])[0] // 1st string is 3-char lang
	src = src[3:]

	if src[0] == 0x00 {
		src = src[1:] // a null separator may appear, skip it
	}

	// Retrieve comment text.
	var texts []string
	if isUtf16 {
		texts = _Util.ConvertUtf16Strings(src)
	} else {
		texts = _Util.ConvertAsciiStrings(src)
	}

	if len(texts) > 1 {
		msg := "Comment has more than 1 text field"
		for _, t := range texts {
			msg += fmt.Sprintf(", \"%s\"", t)
		}
		return nil, errors.New(msg)
	}

	fr.text = texts[0]
	return fr, nil
}

func (me *_Parser) parseTextFrame(src []byte) ([]string, error) {
	switch src[0] {
	case 0x00:
		// Encoding is ISO-8859-1.
		return _Util.ConvertAsciiStrings(src[1:]), nil // skip 0x00 encoding byte
	case 0x01:
		// Encoding is Unicode UTF-16, may have 2-byte BOM.
		return _Util.ConvertUtf16Strings(src[1:]), nil // skip 0x01 encoding byte
	default:
		return nil, errors.New(
			fmt.Sprintf("Text frame with unknown text encoding (%d).", src[0]),
		)
	}
}

func (me *_Parser) parseBinaryFrame(src []byte) *FrameBinary {
	fr := &FrameBinary{}
	fr.data = make([]byte, len(src))
	copy(fr.data, src) // simply store bytes
	return fr
}
