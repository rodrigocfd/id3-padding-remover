package id3

import (
	"encoding/binary"
	"errors"
)

type FRAME uint8

const (
	FRAME_UNDEFINED FRAME = iota
	FRAME_TEXT
	FRAME_BINARY
)

type Frame struct {
	frameSize uint32
	name4     string
	kind      FRAME
	texts     []string
	binData   []byte
}

func (me *Frame) Read(src []byte) error {
	me.frameSize = binary.BigEndian.Uint32(src[4:8]) + 10 // also count 10-byte tag header
	me.name4 = string(src[0:4])

	if me.name4[0] == 'T' || me.name4 == "COMM" {
		me.kind = FRAME_TEXT

		if src[10] == 0x00 { // encoding is ISO-8859-1
			me.parseAscii(src[10:])
		} else if src[10] == 0x01 { // encoding is Unicode UTF-16 with 2-byte BOM
			me.parseUtf16(src[10:])
		} else {
			return errors.New("Unknown text encoding.")
		}

	} else {
		me.kind = FRAME_BINARY
		copy(me.binData, src[10:10+me.frameSize]) // simply store bytes
	}

	return nil
}

func (me *Frame) parseAscii(src []byte) {
	frameDataSize := me.frameSize - 10 // minus header size

	off := 1 // skip encoding byte
	offBase := 1

	// Parse any number of null-separated strings.
	for {
		if src[off] == 0x00 || uint32(off) == frameDataSize-1 { // we reached the end of a string
			me.texts = append(me.texts, string(src[offBase:off+1]))
			offBase = off + 1
		}
		off++
		if uint32(off) == frameDataSize { // end of frame
			break
		}
	}
}

func (me *Frame) parseUtf16(src []byte) {

}
