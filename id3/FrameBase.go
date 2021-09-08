package id3

import (
	"encoding/binary"
	"id3fit/id3/util"
)

type _FrameBase struct {
	name4        string
	originalSize int
}

func (me *_FrameBase) parse(src []byte) {
	me.name4 = string(src[0:4])
	me.originalSize = int(binary.BigEndian.Uint32(src[4:8]) + 10) // also count 10-byte tag header
}

func (me *_FrameBase) Name4() string     { return me.name4 }
func (me *_FrameBase) OriginalSize() int { return me.originalSize }

func (me *_FrameBase) serializeHeader(totalFrameSize int) []byte {
	blob := make([]byte, 0, 10) // header is 10 bytes
	blob = append(blob, []byte(me.name4)...)

	blob = util.Append32(blob, binary.BigEndian, uint32(totalFrameSize-10)) // without 10-byte header

	blob = util.Append16(blob, binary.BigEndian, 0x0000) // flags
	return blob
}
