use super::Tag;

/// MP3 file path and its ID3v2 tag.
pub struct PathAndTag {
	pub mp3_path: String,
	pub tag: Tag,
}

impl PathAndTag {
	#[must_use]
	pub fn new(mp3_path: &str, tag: Tag) -> Self {
		Self {
			mp3_path: mp3_path.to_owned(),
			tag,
		}
	}
}
