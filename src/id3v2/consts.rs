/// Known tag fields.
#[derive(Clone, Copy)]
pub enum Field {
	Artist,
	Title,
	Album,
	Track,
	Year,
	Genre,
}

impl Field {
	/// Returns the raw 4-char identifier of the frame.
	#[must_use]
	pub const fn name4(&self) -> &str {
		use Field as F;
		match self {
			F::Artist => "TPE1",
			F::Title => "TIT2",
			F::Album => "TALB",
			F::Track => "TRCK",
			F::Year => "TYER",
			F::Genre => "TCON",
		}
	}
}

/// APIC picture types.
#[derive(Clone, Copy)]
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
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		use PicType as P;
		write!(f, "{}", match self {
			P::Other => "Other",
			P::FileIconPng32 => "32x32 pixels 'file icon' (PNG only)",
			P::FileIconOther => "Other file icon",
			P::CoverFront => "Cover (front)",
			P::CoverBack => "Cover (back)",
			P::Leaflet => "Leaflet page",
			P::CdLabelSide => "Media (e.g. lable side of CD)",
			P::LeadArtist => "Lead artist/lead performer/soloist",
			P::Artist => "Artist/performer",
			P::Conductor => "Conductor",
			P::Band => "Band/Orchestra",
			P::Composer => "Composer",
			P::Lyricist => "Lyricist/text writer",
			P::RecLocation => "Recording Location",
			P::DuringRecording => "During recording",
			P::DuringPerformance => "During performance",
			P::MovieCapture => "Movie/video screen capture",
			P::BrightColouredFish => "A bright coloured fish",
			P::Illustration => "Illustration",
			P::BandLogo => "Band/artist logotype",
			P::PublisherLogo => "Publisher/Studio logotype",
		})
	}
}

impl PicType {
	#[must_use]
	pub fn from_u8(n: u8) -> Self {
		if n > 0x14 {
			panic!("Invalid PicType value: {}.", n);
		}
		unsafe { std::mem::transmute(n) }
	}
}
