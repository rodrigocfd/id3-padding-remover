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
	SerializeSize() int
	Serialize(dest []byte) []byte
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

// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

// Concrete type for [text information] data.
//
// [text information]: https://id3.org/id3v2.3.0#Text_information_frames
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

func (me *BodyText) SerializeSize() int {
	encByte := serializeEnc(me.Text)
	szText, _ := serializeStrSize(encByte, me.Text)
	return 1 + szText
}

func (me *BodyText) Serialize(dest []byte) []byte {
	encByte := serializeEnc(me.Text)
	dest = append(dest, byte(encByte))
	dest = serializeStr(encByte, dest, me.Text)
	return dest
}

// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

// Concrete type for [user-defined text] data.
//
// [user-defined text]: https://id3.org/id3v2.3.0#User_defined_text_information_frame
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

func (me *BodyUserText) SerializeSize() int {
	encByte := serializeEnc(me.Descr, me.Text)
	szDescr, _ := serializeStrSize(encByte, me.Descr)
	szText, _ := serializeStrSize(encByte, me.Text)
	return 1 + szDescr + szText
}

func (me *BodyUserText) Serialize(dest []byte) []byte {
	encByte := serializeEnc(me.Descr, me.Text)
	dest = append(dest, byte(encByte))
	dest = serializeStr(encByte, dest, me.Descr)
	dest = serializeStr(encByte, dest, me.Text)
	return dest
}

// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

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

func (me *BodyBinary) SerializeSize() int {
	return len(me.Bin)
}

func (me *BodyBinary) Serialize(dest []byte) []byte {
	dest = append(dest, me.Bin...)
	return dest
}

// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

// Concrete type for [comment] data.
//
// [comment]: https://id3.org/id3v2.3.0#Comments
type BodyComment struct {
	Lang3 string
	Descr string
	Text  string
}

// Constructor.
func _BodyCommentParse(src []byte) (*BodyComment, error) {
	encByte, src, err := parseEnc(src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyComment: %w", err)
	}

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

func (me *BodyComment) SerializeSize() int {
	encByte := serializeEnc(me.Descr, me.Text)
	szDescr, _ := serializeStrSize(encByte, me.Descr)
	szText, _ := serializeStrSize(encByte, me.Text)
	return 1 + 3 + szDescr + szText
}

func (me *BodyComment) Serialize(dest []byte) []byte {
	encByte := serializeEnc(me.Descr, me.Text)
	dest = append(dest, byte(encByte))
	dest = append(dest, []byte(me.Lang3)...)
	dest = serializeStr(encByte, dest, me.Descr)
	dest = serializeStr(encByte, dest, me.Text)
	return dest
}

// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

// Concrete type for [picture] frame.
//
// [picture]: https://id3.org/id3v2.3.0#Attached_picture
type BodyPicture struct {
	Mime  string
	Type  PICTYPE
	Descr string
	Bin   []byte
}

// Constructor.
func _BodyPictureParse(src []byte) (*BodyPicture, error) {
	encByte, src, err := parseEnc(src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyPicture: %w", err)
	}

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

func (me *BodyPicture) SerializeSize() int {
	encByte := serializeEnc(me.Descr)
	szMime, _ := serializeStrSize(ENC_ISO88591, me.Mime)
	szDescr, _ := serializeStrSize(encByte, me.Descr)
	return 1 + szMime + 1 + szDescr + len(me.Bin)
}

func (me *BodyPicture) Serialize(dest []byte) []byte {
	encByte := serializeEnc(me.Descr)
	dest = append(dest, byte(encByte))
	dest = serializeStr(ENC_ISO88591, dest, me.Mime)
	dest = append(dest, byte(me.Type))
	dest = serializeStr(encByte, dest, me.Descr)
	dest = append(dest, me.Bin...)
	return dest
}

// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

// Concrete type for [general encapsulated object] data.
//
// [general encapsulated object]: https://id3.org/id3v2.3.0#General_encapsulated_object
type BodyGeob struct {
	Mime     string
	FileName string
	Descr    string
	EncObj   []byte
}

// Constructor.
func _BodyGeobParse(src []byte) (*BodyGeob, error) {
	encByte, src, err := parseEnc(src)
	if err != nil {
		return nil, fmt.Errorf("parse BodyGeob: %w", err)
	}

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

func (me *BodyGeob) SerializeSize() int {
	encByte := serializeEnc(me.FileName, me.Descr)
	szMime, _ := serializeStrSize(ENC_ISO88591, me.Mime)
	szFileName, _ := serializeStrSize(encByte, me.FileName)
	szDescr, _ := serializeStrSize(encByte, me.Descr)
	return 1 + szMime + szFileName + szDescr + len(me.EncObj)
}

func (me *BodyGeob) Serialize(dest []byte) []byte {
	encByte := serializeEnc(me.FileName, me.Descr)
	dest = append(dest, byte(encByte))
	dest = serializeStr(ENC_ISO88591, dest, me.Mime)
	dest = serializeStr(encByte, dest, me.FileName)
	dest = serializeStr(encByte, dest, me.Descr)
	dest = append(dest, me.EncObj...)
	return dest
}
