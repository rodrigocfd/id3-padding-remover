package id3

type Frame interface {
	FrameSize() uint32
	Name4() string
}

type baseFrame struct {
	name4     string
	frameSize uint32
}

func (me *baseFrame) Name4() string     { return me.name4 }
func (me *baseFrame) FrameSize() uint32 { return me.frameSize }

// Contains binary data, no further validations performed.
type FrameBinary struct {
	baseFrame
	data []byte
}

func (me *FrameBinary) Data() []byte { return me.data }

// Contains one single text field.
type FrameText struct {
	baseFrame
	text string
}

func (me *FrameText) Text() string { return me.text }

// Contains many text fields.
type FrameMultiText struct {
	baseFrame
	texts []string
}

func (me *FrameMultiText) Texts() []string { return me.texts }

// Commentary is a special case of multi text.
type FrameComment struct {
	baseFrame
	lang string
	text string
}

func (me *FrameComment) Lang() string { return me.lang }
func (me *FrameComment) Text() string { return me.text }
