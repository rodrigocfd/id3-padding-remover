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
	FRAMETXT_PERFORMER   FRAMETXT = "TPE3"
	FRAMETXT_COMPOSER    FRAMETXT = "TCOM"
	FRAMETXT_LYRICIST    FRAMETXT = "TEXT"
	FRAMETXT_PUBLISHER   FRAMETXT = "TPUB"
	FRAMETXT_ORIG_ARTIST FRAMETXT = "TOPE"
	FRAMETXT_ORIG_ALBUM  FRAMETXT = "TOAL"
	FRAMETXT_ORIG_YEAR   FRAMETXT = "TORY"
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

var _TABLE_PICTYPE_STR map[PICTYPE]string = map[PICTYPE]string{
	PICTYPE_OTHER:                "Other",
	PICTYPE_FILE_ICON_PNG_32:     "32x32 pixels 'file icon' (PNG only)",
	PICTYPE_FILE_ICON_OTHER:      "Other file icon",
	PICTYPE_COVER_FRONT:          "Cover (front)",
	PICTYPE_COVER_BACK:           "Cover (back)",
	PICTYPE_LEAFLET:              "Leaflet page",
	PICTYPE_CD_LABEL_SIDE:        "Media (e.g. lable side of CD)",
	PICTYPE_LEAD_ARTIST:          "Lead artist/lead performer/soloist",
	PICTYPE_ARTIST:               "Artist/performer",
	PICTYPE_CONDUCTOR:            "Conductor",
	PICTYPE_BAND:                 "Band/Orchestra",
	PICTYPE_COMPOSER:             "Composer",
	PICTYPE_LYRICIST:             "Lyricist/text writer",
	PICTYPE_REC_LOCATION:         "Recording Location",
	PICTYPE_DURING_RECORDING:     "During recording",
	PICTYPE_DURING_PERFORMANCE:   "During performance",
	PICTYPE_MOVIE_CAPTURE:        "Movie/video screen capture",
	PICTYPE_BRIGHT_COLOURED_FISH: "A bright coloured fish",
	PICTYPE_ILLUSTRATION:         "Illustration",
	PICTYPE_BAND_LOGO:            "Band/artist logotype",
	PICTYPE_PUBLISHER_LOGO:       "Publisher/Studio logotype",
}

// Returns the descriptive name of the PICTYPE.
func (p PICTYPE) String() string {
	if descr, ok := _TABLE_PICTYPE_STR[p]; ok {
		return descr
	} else {
		panic(fmt.Sprintf("Invalid PICTYPE: %d.", p))
	}
}
