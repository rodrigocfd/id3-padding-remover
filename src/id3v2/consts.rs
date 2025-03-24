/// Encoding byte.
#[derive(Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Enc {
	Iso88591 = 0x00,
	Unicode = 0x01,
}

impl TryFrom<u8> for Enc {
	type Error = String;

	fn try_from(n: u8) -> Result<Self, Self::Error> {
		match n {
			0 => Ok(Self::Iso88591),
			1 => Ok(Self::Unicode),
			n => Err(format!("Invalid encoding byte: {n}.")),
		}
	}
}

impl From<Enc> for u8 {
	fn from(enc: Enc) -> Self {
		match enc {
			Enc::Iso88591 => 0x00,
			Enc::Unicode => 0x01,
		}
	}
}

/// APIC picture types.
#[derive(Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
#[allow(unused)]
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

impl From<u8> for PicType {
	fn from(n: u8) -> Self {
		if n > 0x14 {
			panic!("Invalid PicType value: {}.", n);
		}
		unsafe { std::mem::transmute(n) }
	}
}
