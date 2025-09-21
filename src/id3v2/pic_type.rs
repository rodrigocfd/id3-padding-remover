/// APIC picture type.
#[derive(Clone, Copy, PartialEq, Eq)]
#[allow(unused)]
#[repr(u8)]
pub enum PicType {
	Other = 0x00,
	FileIconPng32 = 0x01,
	FileIconOther = 0x02,
	CoverFront = 0x03,
	CoverBack = 0x04,
	Leaflet = 0x05,
	CdLabelSide = 0x06,
	LeadArtist = 0x07,
	Artist = 0x08,
	Conductor = 0x09,
	Band = 0x0a,
	Composer = 0x0b,
	Lyricist = 0x0c,
	RecLocation = 0x0d,
	DuringRecording = 0x0e,
	DuringPerformance = 0x0f,
	MovieCapture = 0x10,
	BrightColouredFish = 0x11,
	Illustration = 0x12,
	BandLogo = 0x13,
	PublisherLogo = 0x14,
}

impl std::fmt::Display for PicType {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		use PicType::*;
		write!(
			f,
			"{}",
			match self {
				Other => "Other",
				FileIconPng32 => "32x32 pixels 'file icon' (PNG only)",
				FileIconOther => "Other file icon",
				CoverFront => "Cover (front)",
				CoverBack => "Cover (back)",
				Leaflet => "Leaflet page",
				CdLabelSide => "Media (e.g. label side of CD)",
				LeadArtist => "Lead artist/lead performer/soloist",
				Artist => "Artist/performer",
				Conductor => "Conductor",
				Band => "Band/Orchestra",
				Composer => "Composer",
				Lyricist => "Lyricist/text writer",
				RecLocation => "Recording Location",
				DuringRecording => "During recording",
				DuringPerformance => "During performance",
				MovieCapture => "Movie/video screen capture",
				BrightColouredFish => "A bright coloured fish",
				Illustration => "Illustration",
				BandLogo => "Band/artist logotype",
				PublisherLogo => "Publisher/Studio logotype",
			}
		)
	}
}

impl TryFrom<u8> for PicType {
	type Error = String;

	fn try_from(n: u8) -> Result<Self, Self::Error> {
		if n <= 0x14 {
			Ok(unsafe { std::mem::transmute(n) })
		} else {
			Err(format!("Invalid pic type byte: {n}."))
		}
	}
}
