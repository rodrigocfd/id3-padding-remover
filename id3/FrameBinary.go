package id3

type FrameBinary struct {
	_FrameBase
	binData []byte
}

// Constructor.
func _ParseFrameBinary(base _FrameBase, src []byte) (*FrameBinary, error) {
	theData := make([]byte, len(src))
	copy(theData, src) // simply store bytes

	return &FrameBinary{
		_FrameBase: base,
		binData:    theData,
	}, nil
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
