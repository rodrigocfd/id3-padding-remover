//go:build windows

package id3v2

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/wstr"
	"github.com/rodrigocfd/xslices"
)

// Polymorphic data of a frame.
type Body interface {
	implBody()
	Clone() Body
	AsText() string
	ForceText(text string)
	Serialize(dest *win.Vec[byte]) int
}

// Constructor.
func _BodyParse(name4 string, src []byte) (Body, error) {
	if name4 == "COMM" {
		comm, err := _BodyCommentParse(src)
		if err != nil {
			return nil, err
		}
		return comm, nil
	} else if name4 == "APIC" {
		apic, err := _BodyPictureParse(src)
		if err != nil {
			return nil, err
		}
		return apic, nil
	} else if name4[0] == 'T' {
		texts, err := parseStrings(src)
		if err != nil {
			return nil, fmt.Errorf("frame %s with bad strings: %w", name4, err)
		}

		switch len(texts) {
		case 0:
			return nil, fmt.Errorf("frame %s contains no texts", name4)
		case 1:
			return &BodyText{Text: texts[0]}, nil
		case 2:
			return &BodyUserText{Descr: texts[0], Text: texts[1]}, nil
		default:
			return nil, fmt.Errorf("frame %s contains %d texts", name4, len(texts))
		}
	} else { // everything else is treated as raw binary
		return _BodyBinaryParse(src), nil
	}
}

// Concrete type.
type BodyText struct {
	Text string
}

func (*BodyText) implBody() {}

func (me *BodyText) Clone() Body {
	return &BodyText{me.Text}
}

func (me *BodyText) AsText() string {
	return me.Text
}

func (me *BodyText) ForceText(text string) {
	me.Text = text
}

func (me *BodyText) Serialize(dest *win.Vec[byte]) int {
	encByte, serializedText := serializeStrings(me.Text)
	packLen := 1 + len(serializedText)

	dest.Reserve(dest.Len() + packLen)
	dest.Append(byte(encByte))
	dest.Append(serializedText...)

	return packLen
}

// Concrete type.
type BodyUserText struct {
	Descr string
	Text  string
}

func (*BodyUserText) implBody() {}

func (me *BodyUserText) Clone() Body {
	return &BodyUserText{me.Descr, me.Text}
}

func (me *BodyUserText) AsText() string {
	if me.Descr != "" {
		return me.Descr + " " + me.Text
	} else {
		return me.Text
	}
}

func (me *BodyUserText) ForceText(text string) {
	me.Descr = ""
	me.Text = text
}

func (me *BodyUserText) Serialize(dest *win.Vec[byte]) int {
	encByte, serializedStrs := serializeStrings(me.Descr, me.Text)
	packLen := 1 + len(serializedStrs)

	dest.Reserve(dest.Len() + packLen)
	dest.Append(byte(encByte))
	dest.Append(serializedStrs...)

	return packLen
}

// Concrete type.
type BodyBinary struct {
	Bin []byte
}

// Constructor.
func _BodyBinaryParse(src []byte) *BodyBinary {
	return &BodyBinary{xslices.ShallowClone(src)} // simply copy all the data
}

func (*BodyBinary) implBody() {}

func (me *BodyBinary) Clone() Body {
	return &BodyBinary{xslices.ShallowClone(me.Bin)}
}

func (me *BodyBinary) AsText() string {
	return wstr.FmtBytes(len(me.Bin))
}

func (me *BodyBinary) ForceText(text string) {
	panic("Cannot set text to a binary frame.")
}

func (me *BodyBinary) Serialize(dest *win.Vec[byte]) int {
	dest.Append(me.Bin...)
	return len(me.Bin)
}

// Concrete type.
type BodyComment struct {
	Lang3 string
	Descr string
	Text  string
}

// Constructor.
func _BodyCommentParse(src []byte) (*BodyComment, error) {
	encByte := ENC(src[0])
	if encByte != ENC_ISO88591 && encByte != ENC_UNICODE {
		return nil, fmt.Errorf("unknown comment encoding: %d", encByte)
	}
	src = src[1:] // skip encoding byte

	me := BodyComment{
		Lang3: string(src[:3]),
	}
	src = src[3:] // skip lang chars

	texts, err := parseStrings(src)
	if err != nil {
		return nil, err
	}

	switch len(texts) {
	case 0:
		return nil, errors.New("comment frame has no texts")
	case 1:
		me.Text = texts[0] // in case of 1 text, be lenient and assume empty description
	case 2:
		me.Descr = texts[0]
		me.Text = texts[1]
	default:
		return nil, fmt.Errorf("comment frame has %d texts", len(texts))
	}

	return &me, nil
}

func (*BodyComment) implBody() {}

func (me *BodyComment) Clone() Body {
	return &BodyComment{me.Lang3, me.Descr, me.Text}
}

func (me *BodyComment) AsText() string {
	if me.Descr != "" {
		return me.Descr + " " + me.Text
	} else {
		return me.Text
	}
}

func (me *BodyComment) ForceText(text string) {
	me.Lang3 = "eng"
	me.Descr = ""
	me.Text = text
}

func (me *BodyComment) Serialize(dest *win.Vec[byte]) int {
	encByte, serializedStrs := serializeStrings(me.Descr, me.Text)
	packLen := 1 + 3 + len(serializedStrs)

	dest.Reserve(dest.Len() + packLen)
	dest.Append(byte(encByte))
	dest.Append([]byte(me.Lang3)...)
	dest.Append(serializedStrs...)

	return packLen
}

// Concrete type.
type BodyPicture struct {
	Mime  string
	Type  PICTYPE
	Descr string
	Bin   []byte
}

// Constructor.
func _BodyPictureParse(src []byte) (*BodyPicture, error) {
	encByte := ENC(src[0])
	if encByte != ENC_ISO88591 && encByte != ENC_UNICODE {
		return nil, fmt.Errorf("unknown picture encoding: %d", encByte)
	}
	src = src[1:] // skip encoding byte

	var me BodyPicture

	mimeParts := bytes.SplitN(src, []byte{0x00}, 2)
	me.Mime = string(mimeParts[0]) // assume ISO-8859-1 mime
	src = mimeParts[1]

	me.Type = PICTYPE(src[0])
	src = src[1:] // skip picture type byte

	if encByte == ENC_ISO88591 {
		descrParts := bytes.SplitN(src, []byte{0x00}, 2)
		texts := parseIso88591Strings(descrParts[0])
		if len(texts) > 0 { // description may be absent
			me.Descr = texts[0]
		}
		src = descrParts[1]
	} else {
		descrParts := bytes.SplitN(src, []byte{0x00, 0x00}, 2)
		texts := parseUnicodeStrings(descrParts[0])
		if len(texts) > 0 { // description may be absent
			me.Descr = texts[0]
		}
		src = descrParts[1]
	}

	me.Bin = make([]byte, len(src)) // simply copy all the data
	copy(me.Bin, src)
	return &me, nil
}

func (*BodyPicture) implBody() {}

func (me *BodyPicture) Clone() Body {
	return &BodyPicture{me.Mime, me.Type, me.Descr, xslices.ShallowClone(me.Bin)}
}

func (me *BodyPicture) AsText() string {
	return fmt.Sprintf("%s %s %s",
		PICNAMES[me.Type], me.Mime, wstr.FmtBytes(len(me.Bin)))
}

func (me *BodyPicture) ForceText(text string) {
	panic("Cannot set text to a picture frame.")
}

func (me *BodyPicture) Serialize(pDest *win.Vec[byte]) int {
	encByte, serializedDescr := serializeStrings(me.Descr)
	packLen := 1 + len(me.Mime) + 1 + 1 + len(serializedDescr) + len(me.Bin)

	pDest.Reserve(pDest.Len() + packLen)
	pDest.Append(byte(encByte))
	pDest.Append([]byte(me.Mime)...)
	pDest.Append(0x00)
	pDest.Append(byte(me.Type))
	pDest.Append(serializedDescr...)
	pDest.Append(me.Bin...)

	return packLen
}
