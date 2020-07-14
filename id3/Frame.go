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

	if me.name4[0] == 'T' || me.name4 == "COMM" {
		if src[10] == 0x00 { // encoding is ISO-8859-1
			me.parseAscii(src[10:me.frameSize])
		} else if src[10] == 0x01 { // encoding is Unicode UTF-16 with 2-byte BOM
			me.parseUtf16(src[10:me.frameSize])
		} else {
			return errors.New("Unknown text encoding.")
		}

		if len(me.texts) == 1 {
			me.kind = FRAME_KIND_TEXT
		} else {
			me.kind = FRAME_KIND_MULTI_TEXT
		}

	} else {
		dataSlice := src[10:me.frameSize]
		me.binData = make([]byte, len(dataSlice))
		copy(me.binData, dataSlice) // simply store bytes
		me.kind = FRAME_KIND_BINARY
	}

	return nil
}

func (me *Frame) parseAscii(src []byte) {
	src = src[1:] // skip encoding byte

	if src[len(src)-1] == 0x00 { // we have a trailing zero, which is useless
		src = src[:len(src)-1]
	}

	// Parse any number of null-separated strings.
	off := 0
	for {
		if off == len(src)-1 || src[off+1] == 0x00 { // we reached the end of frame contents, or string
			runes := make([]rune, 0, len(src[:off+1]))
			for _, ch := range src[:off+1] {
				runes = append(runes, rune(ch)) // brute force byte to rune
			}
			me.texts = append(me.texts, string(runes)) // then convert from rune slice to string

			if off == len(src)-1 { // no more data
				break
			}
			src = src[off+2:] // skip null separator between strings
			off = 0
		} else {
			off++
		}
	}
}

func (me *Frame) parseUtf16(src []byte) {

}
