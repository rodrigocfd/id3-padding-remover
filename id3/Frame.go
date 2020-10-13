package id3

// Frame is polymorphic, the underlying type will expose the methods to access the contents.
// Note that the methods don't allow the user to change name4/totalFrameSize.
type Frame interface {
	Name4() string
	TotalFrameSize() uint
}

type _BaseFrame struct { // implements Frame
	name4          string
	totalFrameSize uint
}

func (me *_BaseFrame) Name4() string        { return me.name4 }
func (me *_BaseFrame) TotalFrameSize() uint { return me.totalFrameSize }

//------------------------------------------------------------------------------

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
