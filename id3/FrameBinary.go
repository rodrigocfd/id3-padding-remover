package id3

type FrameBinary struct {
	_FrameBase
	binData []byte
}

func (me *FrameBinary) BinData() *[]byte { return &me.binData }

func (me *FrameBinary) parse(base _FrameBase, src []byte) {
	theData := make([]byte, len(src))
	copy(theData, src) // simply store bytes

	me._FrameBase = base
	me.binData = theData
}

func (me *FrameBinary) Serialize() ([]byte, error) {
	totalFrameSize := 10 + len(me.binData) // include header
	header, err := me._FrameBase.serializeHeader(totalFrameSize)
	if err != nil {
		return nil, err
	}

	final := make([]byte, 0, totalFrameSize)
	final = append(final, header...)
	final = append(final, me.binData...)

	return final, nil
}
