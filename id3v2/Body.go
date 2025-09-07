//go:build windows

package id3v2

import (
	"fmt"

	"github.com/rodrigocfd/windigo/wstr"
	"github.com/rodrigocfd/xslices"
)

// Polymorphic data of a frame.
type Body interface {
	implBody()
	Clone() Body
	AsText() string
	ForceText(text string)
	Serialize() []byte
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
	} else if name4 == "GEOB" {
		geob, err := _BodyGeobParse(src)
		if err != nil {
			return nil, err
		}
		return geob, nil
	} else if name4 == "TXXX" {
		ut, err := _BodyUserTextParse(src)
		if err != nil {
			return nil, err
		}
		return ut, nil
	} else if name4[0] == 'T' && name4 != "TDAT" {
		txt, err := _BodyTextParse(src)
		if err != nil {
			return nil, err
		}
		return txt, nil
	} else { // everything else is treated as raw binary
		return _BodyBinaryParse(src), nil
	}
}

// Concrete type.
type BodyText struct {
	Text string
}

// Constructor.
func _BodyTextParse(src []byte) (*BodyText, error) {
	encByte, src, err := parseEnc(src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyText: %w", err)
	}

	info, _, err := parseStr(encByte, src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyText: %w", err)
	}

	return &BodyText{info}, nil
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

func (me *BodyText) Serialize() []byte {
	encByte := serializeEnc(me.Text)
	text := serializeStr(encByte, me.Text)

	szBytes := 1 + len(text)
	blob := make([]byte, 0, szBytes)
	blob = append(blob, byte(encByte))
	blob = append(blob, text...)

	return blob
}

// Concrete type.
type BodyUserText struct {
	Descr string
	Text  string
}

// Constructor.
func _BodyUserTextParse(src []byte) (*BodyUserText, error) {
	encByte, src, err := parseEnc(src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyUserText: %w", err)
	}

	descr, src, err := parseStr(encByte, src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyUserText: %w", err)
	}

	value, _, err := parseStr(encByte, src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyUserText: %w", err)
	}

	return &BodyUserText{descr, value}, nil
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

func (me *BodyUserText) Serialize() []byte {
	encByte := serializeEnc(me.Descr, me.Text)
	descr := serializeStr(encByte, me.Descr)
	text := serializeStr(encByte, me.Text)

	szBytes := 1 + len(descr) + len(text)
	blob := make([]byte, 0, szBytes)
	blob = append(blob, byte(encByte))
	blob = append(blob, descr...)
	blob = append(blob, text...)

	return blob
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

func (me *BodyBinary) Serialize() []byte {
	return xslices.ShallowClone(me.Bin)
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

	lang3 := string(src[:3])
	src = src[3:] // skip lang chars

	descr, src, err := parseStr(encByte, src)
	if err != nil {
		return nil, err
	}

	text, _, err := parseStr(encByte, src)
	if err != nil {
		return nil, err
	}

	return &BodyComment{lang3, descr, text}, nil
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

func (me *BodyComment) Serialize() []byte {
	encByte := serializeEnc(me.Descr, me.Text)
	descr := serializeStr(encByte, me.Descr)
	text := serializeStr(encByte, me.Text)

	szBytes := 1 + 3 + len(descr) + len(text)
	blob := make([]byte, 0, szBytes)
	blob = append(blob, byte(encByte))
	blob = append(blob, []byte(me.Lang3)...)
	blob = append(blob, descr...)
	blob = append(blob, text...)

	return blob
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

	mime, src, _ := parseStr(ENC_ISO88591, src)

	ty := PICTYPE(src[0])
	src = src[1:] // skip picture type byte

	descr, src, err := parseStr(encByte, src)
	if err != nil {
		return nil, err
	}

	bin := xslices.ShallowClone(src) // simply copy all the data

	return &BodyPicture{mime, ty, descr, bin}, nil
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

func (me *BodyPicture) Serialize() []byte {
	encByte := serializeEnc(me.Descr)
	mime := serializeStr(ENC_ISO88591, me.Mime)
	descr := serializeStr(encByte, me.Descr)

	szBytes := 1 + len(mime) + 1 + len(descr) + len(me.Bin)
	blob := make([]byte, 0, szBytes)
	blob = append(blob, byte(encByte))
	blob = append(blob, mime...)
	blob = append(blob, byte(me.Type))
	blob = append(blob, descr...)
	blob = append(blob, me.Bin...)

	return blob
}

// Concrete type.
type BodyGeob struct {
	Mime     string
	FileName string
	Descr    string
	EncObj   []byte
}

// Constructor.
func _BodyGeobParse(src []byte) (*BodyGeob, error) {
	encByte := ENC(src[0])
	if encByte != ENC_ISO88591 && encByte != ENC_UNICODE {
		return nil, fmt.Errorf("unknown general encapsulated object encoding: %d", encByte)
	}
	src = src[1:] // skip encoding byte

	mime, src, _ := parseStr(ENC_ISO88591, src)

	filename, src, err := parseStr(encByte, src)
	if err != nil {
		return nil, err
	}

	descr, src, err := parseStr(encByte, src)
	if err != nil {
		return nil, err
	}

	encObj := xslices.ShallowClone(src)

	return &BodyGeob{mime, filename, descr, encObj}, nil
}

func (*BodyGeob) implBody() {}

func (me *BodyGeob) Clone() Body {
	return &BodyGeob{me.Mime, me.FileName, me.Descr, xslices.ShallowClone(me.EncObj)}
}

func (me *BodyGeob) AsText() string {
	return fmt.Sprintf("%s %s %s",
		me.Mime, me.Descr, wstr.FmtBytes(len(me.EncObj)))
}

func (me *BodyGeob) ForceText(text string) {
	panic("Cannot set text to a general encapsulated object frame.")
}

func (me *BodyGeob) Serialize() []byte {
	encByte := serializeEnc(me.FileName, me.Descr)
	mime := serializeStr(ENC_ISO88591, me.Mime)
	filename := serializeStr(encByte, me.FileName)
	descr := serializeStr(encByte, me.Descr)

	szBlob := 1 + len(mime) + len(filename) + len(descr) + len(me.EncObj)
	blob := make([]byte, 0, szBlob)
	blob = append(blob, byte(encByte))
	blob = append(blob, mime...)
	blob = append(blob, filename...)
	blob = append(blob, descr...)
	blob = append(blob, me.EncObj...)

	return blob
}
