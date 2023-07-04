package id3v2

import (
	"bytes"
	"errors"
	"fmt"
	"id3fit/id3v2/util"
)

// Polymorphic data of a frame.
type FrameData interface {
	implFrameData()
	Serialize() []byte
}

//------------------------------------------------------------------------------

type FrameDataText struct {
	Text string
}

func (*FrameDataText) implFrameData() {}

func (f *FrameDataText) Serialize() []byte {
	encodingByte, serialized := util.SerializeStrings([]string{f.Text})

	buf := make([]byte, 0, 1+len(serialized))
	buf = append(buf, encodingByte)
	buf = append(buf, serialized...)
	return buf
}

//------------------------------------------------------------------------------

type FrameDataUserText struct {
	Descr string
	Text  string
}

func (*FrameDataUserText) implFrameData() {}

func (f *FrameDataUserText) Serialize() []byte {
	encodingByte, serialized := util.SerializeStrings([]string{f.Descr, f.Text})

	buf := make([]byte, 0, 1+len(serialized))
	buf = append(buf, encodingByte)
	buf = append(buf, serialized...)
	return buf
}

//------------------------------------------------------------------------------

type FrameDataBinary struct {
	Data []byte
}

func (*FrameDataBinary) implFrameData() {}

// Constructor.
func _NewFrameDataBinary(src []byte) *FrameDataBinary {
	f := &FrameDataBinary{
		Data: make([]byte, len(src)),
	}
	copy(f.Data, src) // so the original source slice can be freed
	return f
}

func (f *FrameDataBinary) Serialize() []byte {
	return f.Data
}

//------------------------------------------------------------------------------

type FrameDataComment struct {
	Lang3 string
	Descr string
	Text  string
}

func (*FrameDataComment) implFrameData() {}

// Constructor.
func _NewFrameDataComment(src []byte) (*FrameDataComment, error) {
	encodingByte := src[0]
	if encodingByte != 0x00 && encodingByte != 0x01 {
		return nil, fmt.Errorf("unknown encoding: %d", encodingByte)
	}
	src = src[1:] // skip encoding byte

	// Create our frame object.
	f := &FrameDataComment{}

	f.Lang3 = string(src[:3])
	src = src[3:] // skip lang chars

	texts, err := util.ParseAnyStrings(src)
	if err != nil {
		return nil, err
	}

	switch len(texts) {
	case 0:
		return nil, errors.New("comment frame has no texts")
	case 1:
		f.Text = texts[0] // in case of 1 text, be lenient and assume empty description
	case 2:
		f.Descr = texts[0]
		f.Text = texts[1]
	default:
		return nil, fmt.Errorf("comment frame has %d texts", len(texts))
	}

	return f, nil
}

func (f *FrameDataComment) Serialize() []byte {
	encodingByte, serialized := util.SerializeStrings([]string{f.Descr, f.Text})

	buf := make([]byte, 0, 1+3+len(serialized))
	buf = append(buf, encodingByte)
	buf = append(buf, []byte(f.Lang3)...)
	buf = append(buf, serialized...)
	return buf
}

//------------------------------------------------------------------------------

type FrameDataPicture struct {
	Mime  string
	Type  PICTYPE
	Descr string
	Data  []byte
}

func (*FrameDataPicture) implFrameData() {}

// Constructor.
func _NewFrameDataPicture(src []byte) (*FrameDataPicture, error) {
	encodingByte := src[0]
	if encodingByte != 0x00 && encodingByte != 0x01 {
		return nil, fmt.Errorf("unknown encoding: %d", encodingByte)
	}
	src = src[1:] // skip encoding byte

	// Create our frame object.
	f := &FrameDataPicture{}

	mimeParts := bytes.SplitN(src, []byte{0x00}, 2)
	f.Mime = string(mimeParts[0]) // assume ISO-8859-1 mime
	src = mimeParts[1]

	f.Type = PICTYPE(src[0])
	src = src[1:] // skip picture type

	if encodingByte == 0x00 { // ISO-8859-1
		descrParts := bytes.SplitN(src, []byte{0x00}, 2)
		texts := util.ParseIso88591Strings(descrParts[0])
		if len(texts) > 0 { // description may be absent
			f.Descr = texts[0]
		}
		src = descrParts[1]
	} else { // Unicode
		descrParts := bytes.SplitN(src, []byte{0x00, 0x00}, 2)
		texts := util.ParseUnicodeStrings(descrParts[0])
		if len(texts) > 0 { // description may be absent
			f.Descr = texts[0]
		}
		src = descrParts[1]
	}

	f.Data = make([]byte, len(src))
	copy(f.Data, src) // so the original source slice can be freed

	return f, nil
}

func (f *FrameDataPicture) Serialize() []byte {
	encodingByte, serialized := util.SerializeStrings([]string{f.Descr})

	buf := make([]byte, 0, 1+len(f.Mime)+1+1+len(serialized)+len(f.Data))
	buf = append(buf, encodingByte)
	buf = append(buf, []byte(f.Mime)...)
	buf = append(buf, 0x00)
	buf = append(buf, byte(f.Type))
	buf = append(buf, serialized...)
	buf = append(buf, f.Data...)
	return buf
}
