package id3v2

import (
	"fmt"
)

// Known frame text fields.
type FRAMETXT string

const (
	FRAMETXT_ARTIST      FRAMETXT = "TPE1"
	FRAMETXT_TITLE       FRAMETXT = "TIT2"
	FRAMETXT_SUBTITLE    FRAMETXT = "TIT3"
	FRAMETXT_ALBUM       FRAMETXT = "TALB"
	FRAMETXT_TRACK       FRAMETXT = "TRCK"
	FRAMETXT_YEAR        FRAMETXT = "TYER"
	FRAMETXT_GENRE       FRAMETXT = "TCON"
	FRAMETXT_COMPOSER    FRAMETXT = "TCOM"
	FRAMETXT_LYRICIST    FRAMETXT = "TEXT"
	FRAMETXT_ORIG_ARTIST FRAMETXT = "TOPE"
	FRAMETXT_ORIG_ALBUM  FRAMETXT = "TOAL"
	FRAMETXT_ORIG_YEAR   FRAMETXT = "TORY"
	FRAMETXT_PERFORMER   FRAMETXT = "TPE3"
	FRAMETXT_COMMENT     FRAMETXT = "COMM"
)

// APIC picture types.
type PICTYPE uint8

const (
	PICTYPE_OTHER                PICTYPE = 0x00
	PICTYPE_FILE_ICON_PNG_32     PICTYPE = 0x01
	PICTYPE_FILE_ICON_OTHER      PICTYPE = 0x02
	PICTYPE_COVER_FRONT          PICTYPE = 0x03
	PICTYPE_COVER_BACK           PICTYPE = 0x04
	PICTYPE_LEAFLET              PICTYPE = 0x05
	PICTYPE_CD_LABEL_SIDE        PICTYPE = 0x06
	PICTYPE_LEAD_ARTIST          PICTYPE = 0x07
	PICTYPE_ARTIST               PICTYPE = 0x08
	PICTYPE_CONDUCTOR            PICTYPE = 0x09
	PICTYPE_BAND                 PICTYPE = 0x0a
	PICTYPE_COMPOSER             PICTYPE = 0x0b
	PICTYPE_LYRICIST             PICTYPE = 0x0c
	PICTYPE_REC_LOCATION         PICTYPE = 0x0d
	PICTYPE_DURING_RECORDING     PICTYPE = 0x0e
	PICTYPE_DURING_PERFORMANCE   PICTYPE = 0x0f
	PICTYPE_MOVIE_CAPTURE        PICTYPE = 0x10
	PICTYPE_BRIGHT_COLOURED_FISH PICTYPE = 0x11
	PICTYPE_ILLUSTRATION         PICTYPE = 0x12
	PICTYPE_BAND_LOGO            PICTYPE = 0x13
	PICTYPE_PUBLISHER_LOGO       PICTYPE = 0x14
)

func (p PICTYPE) String() string {
	switch p {
	case PICTYPE_OTHER:
		return "Other"
	case PICTYPE_FILE_ICON_PNG_32:
		return "32x32 pixels 'file icon' (PNG only)"
	case PICTYPE_FILE_ICON_OTHER:
		return "Other file icon"
	case PICTYPE_COVER_FRONT:
		return "Cover (front)"
	case PICTYPE_COVER_BACK:
		return "Cover (back)"
	case PICTYPE_LEAFLET:
		return "Leaflet page"
	case PICTYPE_CD_LABEL_SIDE:
		return "Media (e.g. lable side of CD)"
	case PICTYPE_LEAD_ARTIST:
		return "Lead artist/lead performer/soloist"
	case PICTYPE_ARTIST:
		return "Artist/performer"
	case PICTYPE_CONDUCTOR:
		return "Conductor"
	case PICTYPE_BAND:
		return "Band/Orchestra"
	case PICTYPE_COMPOSER:
		return "Composer"
	case PICTYPE_LYRICIST:
		return "Lyricist/text writer"
	case PICTYPE_REC_LOCATION:
		return "Recording Location"
	case PICTYPE_DURING_RECORDING:
		return "During recording"
	case PICTYPE_DURING_PERFORMANCE:
		return "During performance"
	case PICTYPE_MOVIE_CAPTURE:
		return "Movie/video screen capture"
	case PICTYPE_BRIGHT_COLOURED_FISH:
		return "A bright coloured fish"
	case PICTYPE_ILLUSTRATION:
		return "Illustration"
	case PICTYPE_BAND_LOGO:
		return "Band/artist logotype"
	case PICTYPE_PUBLISHER_LOGO:
		return "Publisher/Studio logotype"
	default:
		panic(fmt.Sprintf("Invalid PICTYPE: %d.", p))
	}
}
