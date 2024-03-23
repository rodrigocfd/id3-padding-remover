use winsafe::{self as w};

use super::Tag;

/// MP3 file path and its ID3v2 tag.
pub struct PathAndTag {
	pub mp3_path: String,
	pub tag: Tag,
}

impl PathAndTag {
	/// Reads the tag from an MP3 file.
	#[must_use]
	pub fn read_from_file(mp3_path: &str) -> w::AnyResult<Self> {
		Ok(Self {
			mp3_path: mp3_path.to_owned(),
			tag: Tag::read_from_file(mp3_path)?,
		})
	}

	/// Saves the tag to an MP3 file. If there are no frames, the tag will be
	/// entirely removed from the file.
	pub fn save_to_file(&self) -> w::AnyResult<()> {
		self.tag.save_to_file(&self.mp3_path)
	}
}
