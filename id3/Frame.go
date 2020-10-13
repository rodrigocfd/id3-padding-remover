package id3

type Frame interface {
	Name4() string
	TotalFrameSize() uint
}

type _BaseFrame struct {
	name4          string
	totalFrameSize uint
}

func (me *_BaseFrame) Name4() string        { return me.name4 }
func (me *_BaseFrame) TotalFrameSize() uint { return me.totalFrameSize }

type FrameBinary struct {
	_BaseFrame
	binData []byte
}

func (me *FrameBinary) BinData() []byte { return me.binData }

type FrameText struct {
	_BaseFrame
	text string
}

func (me *FrameText) Text() *string { return &me.text }

type FrameMultiText struct {
	_BaseFrame
	texts []string
}

func (me *FrameMultiText) Texts() []string { return me.texts }

type FrameComment struct {
	_BaseFrame
	lang string
	text string
}

func (me *FrameComment) Lang() *string { return &me.lang }
func (me *FrameComment) Text() *string { return &me.text }
