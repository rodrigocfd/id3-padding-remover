package id3

type Frame interface {
	FrameSize() uint32
	Name4() string
}

type _BaseFrame struct {
	name4     string
	frameSize uint32
}

func (me *_BaseFrame) Name4() string     { return me.name4 }
func (me *_BaseFrame) FrameSize() uint32 { return me.frameSize }

// Contains binary data, no further validations performed.
type FrameBinary struct {
	_BaseFrame
	data []byte
}

func (me *FrameBinary) Data() []byte { return me.data }

// Contains one single text field.
type FrameText struct {
	_BaseFrame
	text string
}

func (me *FrameText) Text() string { return me.text }

// Contains many text fields.
type FrameMultiText struct {
	_BaseFrame
	texts []string
}

func (me *FrameMultiText) Texts() []string { return me.texts }

// Commentary is a special case of multi text.
type FrameComment struct {
	_BaseFrame
	lang string
	text string
}

func (me *FrameComment) Lang() string { return me.lang }
func (me *FrameComment) Text() string { return me.text }
