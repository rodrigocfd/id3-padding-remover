package id3v2

type ErrorNoTagFound struct{}

func (e *ErrorNoTagFound) Error() string {
	return "no ID3 tag found"
}
