use winsafe::{self as w};

use super::frame_data::FrameData;
use super::str_engine;

/// A unit of data within a tag.
#[derive(PartialEq, Eq)]
pub struct Frame {
	name4: String,
	flags: (u8, u8),
	data: FrameData,
}

impl std::fmt::Display for Frame {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		write!(f, "{}: {}", self.name4, self.data)
	}
}

impl Frame {
	#[must_use]
	pub(in crate::id3v2) fn new_from_string(name4: &str, text: &str) -> w::AnyResult<Self> {
		Ok(Self {
			name4: name4.to_owned(),
			flags: (0, 0),
			data: FrameData::new_from_string(name4, text)?,
		})
	}

	/// Also returns declared size, including 10-byte frame header.
	#[must_use]
	pub(in crate::id3v2) fn parse(src: &[u8]) -> w::AnyResult<(Self, u32)> {
		let mut src = src;

		// Parse the 10-byte frame header.
		let name4 = str_engine::from_ascii(&src[0..4]);
		let original_size = u32::from_be_bytes(src[4..8].try_into()?) + 10; // also count 10-byte frame header
		let flags = (src[8], src[9]);

		// Skip frame header, truncate to frame size.
		src = &src[10..original_size as usize];

		// Parse the frame contents.
		let data = FrameData::parse(&name4, src)?;

		Ok((Self { name4, flags, data }, original_size))
	}

	#[must_use]
	pub(in crate::id3v2) fn serialize(&self) -> Vec<u8> {
		let serialized_data = self.data.serialize();
		str_engine::to_ascii(&self.name4).into_iter()
			.chain((serialized_data.len() as u32).to_be_bytes()) // won't count 10-byte header
			.chain([self.flags.0, self.flags.1].into_iter())
			.chain(serialized_data.into_iter())
			.collect()
	}

	#[must_use]
	pub const fn name4(&self) -> &String {
		&self.name4
	}

	#[must_use]
	pub const fn data(&self) -> &FrameData {
		&self.data
	}

	pub fn set_string(&mut self, val: &str) -> w::AnyResult<()> {
		self.data.set_string(val)
	}

	#[must_use]
	pub fn is_replay_gain(&self) -> bool {
		if self.name4 == "TXXX" {
			if let FrameData::UserText(f) = &self.data {
				return f.descr.starts_with("replaygain_");
			}
		}
		false
	}
}
