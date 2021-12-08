/// APIC kind, the type of picture.
#[derive(Clone, Copy, PartialEq, Eq)]
pub enum PicKind {
	Other = 0x00,
	FileIcon32x32,
	OtherFileIcon,
	CoverFront,
	CoverBack,
	LeafletPage,
	Media,
	LeadArtist,
	Artist,
	Conductor,
	Band,
	Composer,
	Lyricist,
	RecordingLocation,
	DuringRecording,
	DuringPerformance,
	ScreenCapture,
	BrightColouredFish,
	Illustration,
	BandLogo,
	PublisherLogo,
}

impl From<u8> for PicKind {
	fn from(v: u8) -> Self {
		match v {
			0x00 => Self::Other,
			0x01 => Self::FileIcon32x32,
			0x02 => Self::OtherFileIcon,
			0x03 => Self::CoverFront,
			0x04 => Self::CoverBack,
			0x05 => Self::LeafletPage,
			0x06 => Self::Media,
			0x07 => Self::LeadArtist,
			0x08 => Self::Artist,
			0x09 => Self::Conductor,
			0x0a => Self::Band,
			0x0b => Self::Composer,
			0x0c => Self::Lyricist,
			0x0d => Self::RecordingLocation,
			0x0e => Self::DuringRecording,
			0x0f => Self::DuringPerformance,
			0x10 => Self::ScreenCapture,
			0x11 => Self::BrightColouredFish,
			0x12 => Self::Illustration,
			0x13 => Self::BandLogo,
			0x14 => Self::PublisherLogo,
			_ => panic!("Invalid picture type."),
		}
	}
}

impl PicKind {
	pub const fn descr(self) -> &'static str {
		match self {
			Self::Other => "Other",
			Self::FileIcon32x32 => "32x32 pixels 'file icon' (PNG only)",
			Self::OtherFileIcon => "Other file icon",
			Self::CoverFront => "Cover (front)",
			Self::CoverBack => "Cover (back)",
			Self::LeafletPage => "Leaflet page",
			Self::Media => "Media (e.g. lable side of CD)",
			Self::LeadArtist => "Lead artist/lead performer/soloist",
			Self::Artist => "Artist/performer",
			Self::Conductor => "Conductor",
			Self::Band => "Band/Orchestra",
			Self::Composer => "Composer",
			Self::Lyricist => "Lyricist/text writer",
			Self::RecordingLocation => "Recording Location",
			Self::DuringRecording => "During recording",
			Self::DuringPerformance => "During performance",
			Self::ScreenCapture => "Movie/video screen capture",
			Self::BrightColouredFish => "A bright coloured fish",
			Self::Illustration => "Illustration",
			Self::BandLogo => "Band/artist logotype",
			Self::PublisherLogo => "Publisher/Studio logotype",
		}
	}
}
