package id3

import (
	"encoding/binary"
	"errors"
)

type FRAME_KIND uint8

const (
	FRAME_KIND_UNDEFINED FRAME_KIND = iota
	FRAME_KIND_TEXT
	FRAME_KIND_MULTI_TEXT
	FRAME_KIND_COMMENT
	FRAME_KIND_BINARY
)

type Frame struct {
	frameSize uint32 // including 10-byte header
	name4     string
	kind      FRAME_KIND
	texts     []string
	binData   []byte
}

func (me *Frame) Name4() string    { return me.name4 }
func (me *Frame) Kind() FRAME_KIND { return me.kind }
func (me *Frame) Texts() []string  { return me.texts }
func (me *Frame) BinData() []byte  { return me.binData }

func (me *Frame) Read(src []byte) error {
	me.frameSize = binary.BigEndian.Uint32(src[4:8]) + 10 // also count 10-byte tag header
	me.name4 = string(src[0:4])
	src = src[10:me.frameSize] // skip frame header, limit to frame size

	if me.name4 == "COMM" {
		return me.parseCommentFrame(src)
	} else if me.name4[0] == 'T' {
		return me.parseTextFrame(src)
	}
	return me.parseBinaryFrame(src) // anything else will be treated as binary
}

func (me *Frame) parseCommentFrame(src []byte) error {
	if src[0] != 0x00 && src[0] != 0x01 {
		return errors.New("Unknown comment encoding.")
	}
	isUtf16 := src[0] == 0x01
	src = src[1:] // skip encoding byte

	me.texts = append(me.texts, convertAsciiStrings(src[:3])[0]) // 1st string is 3-char lang
	src = src[3:]

	if src[0] == 0x00 {
		src = src[1:] // a null separator may appear, skip it
	}

	if isUtf16 {
		me.texts = append(me.texts, convertUtf16Strings(src)...)
	} else {
		me.texts = append(me.texts, convertAsciiStrings(src)...)
	}

	me.kind = FRAME_KIND_COMMENT
	return nil
}

func (me *Frame) parseTextFrame(src []byte) error {
	switch src[0] {
	case 0x00:
		// Encoding is ISO-8859-1.
		me.texts = append(me.texts, convertAsciiStrings(src[1:])...) // skip 0x00 encoding byte
	case 0x01:
		// Encoding is Unicode UTF-16, may have 2-byte BOM.
		me.texts = append(me.texts, convertUtf16Strings(src[1:])...) // skip 0x01 encoding byte
	default:
		return errors.New("Unknown text encoding.")
	}

	if len(me.texts) == 1 {
		me.kind = FRAME_KIND_TEXT
	} else {
		me.kind = FRAME_KIND_MULTI_TEXT // usually TXXX frames with ReplayGain info
	}
	return nil
}

func (me *Frame) parseBinaryFrame(src []byte) error {
	me.binData = make([]byte, len(src))
	copy(me.binData, src) // simply store bytes
	me.kind = FRAME_KIND_BINARY
	return nil
}
