//go:build windows

package id3v2

// APIC picture types.
type PICTYPE uint8

const (
	PICTYPE_OTHER PICTYPE = iota + 0x00
	PICTYPE_FILE_ICON_PNG_32
	PICTYPE_FILE_ICON_OTHER
	PICTYPE_COVER_FRONT
	PICTYPE_COVER_BACK
	PICTYPE_LEAFLET
	PICTYPE_CD_LABEL_SIDE
	PICTYPE_LEAD_ARTIST
	PICTYPE_ARTIST
	PICTYPE_CONDUCTOR
	PICTYPE_BAND
	PICTYPE_COMPOSER
	PICTYPE_LYRICIST
	PICTYPE_REC_LOCATION
	PICTYPE_DURING_RECORDING
	PICTYPE_DURING_PERFORMANCE
	PICTYPE_MOVIE_CAPTURE
	PICTYPE_BRIGHT_COLOURED_FISH
	PICTYPE_ILLUSTRATION
	PICTYPE_BAND_LOGO
	PICTYPE_PUBLISHER_LOGO
)

var PICNAMES = []string{ // index matches PICTYPE value
	"Other",
	"32x32 pixels 'file icon' (PNG only)",
	"Other file icon",
	"Cover (front)",
	"Cover (back)",
	"Leaflet page",
	"Media (e.g. label side of CD)",
	"Lead artist/lead performer/soloist",
	"Artist/performer",
	"Conductor",
	"Band/Orchestra",
	"Composer",
	"Lyricist/text writer",
	"Recording Location",
	"During recording",
	"During performance",
	"Movie/video screen capture",
	"A bright coloured fish",
	"Illustration",
	"Band/artist logotype",
	"Publisher/Studio logotype",
}
