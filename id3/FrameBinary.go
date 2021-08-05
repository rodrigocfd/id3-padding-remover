package id3

type FrameBinary struct {
	_FrameBase
	binData []byte
}

func (me *FrameBinary) parse(base *_FrameBase, src []byte) {
	theData := make([]byte, len(src))
	copy(theData, src) // simply store bytes

	me._FrameBase = *base
	me.binData = theData
}

func (me *FrameBinary) BinData() *[]byte { return &me.binData }

func (me *FrameBinary) Serialize() []byte {
	totalFrameSize := 10 + len(me.binData) // header
	header := me._FrameBase.serializeHeader(totalFrameSize)

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)
	final = append(final, me.binData...)

	return final
}
