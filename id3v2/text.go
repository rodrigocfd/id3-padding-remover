package id3v2

// Names of accepted text fields.
type TEXT string

const (
	TEXT_ARTIST   TEXT = "TPE1"
	TEXT_TITLE    TEXT = "TIT2"
	TEXT_ALBUM    TEXT = "TALB"
	TEXT_TRACK    TEXT = "TRCK"
	TEXT_YEAR     TEXT = "TYER"
	TEXT_GENRE    TEXT = "TCON"
	TEXT_COMPOSER TEXT = "TCOM"
	TEXT_COMMENT  TEXT = "COMM"
)

// Returns a list of all accepted text field constants.
func TextFieldConsts() []TEXT {
	// Note: This must be in sync with dlgfields.checksAndInputs().
	return []TEXT{TEXT_ARTIST, TEXT_TITLE, TEXT_ALBUM, TEXT_TRACK,
		TEXT_YEAR, TEXT_GENRE, TEXT_COMPOSER, TEXT_COMMENT}
}
