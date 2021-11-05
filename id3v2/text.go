package id3v2

// Names of accepted text fields.
type TEXT string

const (
	TEXT_ARTIST      TEXT = "TPE1"
	TEXT_TITLE       TEXT = "TIT2"
	TEXT_SUBTITLE    TEXT = "TIT3"
	TEXT_ALBUM       TEXT = "TALB"
	TEXT_TRACK       TEXT = "TRCK"
	TEXT_YEAR        TEXT = "TYER"
	TEXT_GENRE       TEXT = "TCON"
	TEXT_COMPOSER    TEXT = "TCOM"
	TEXT_LYRICIST    TEXT = "TEXT"
	TEXT_ORIG_ARTIST TEXT = "TOPE"
	TEXT_ORIG_ALBUM  TEXT = "TOAL"
	TEXT_ORIG_YEAR   TEXT = "TORY"
	TEXT_PERFORMER   TEXT = "TPE3"
	TEXT_COMMENT     TEXT = "COMM"
)
