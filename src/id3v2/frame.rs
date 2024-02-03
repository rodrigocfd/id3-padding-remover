use winsafe::{self as w};

use super::frame_data::FrameData;
use super::str_engine;

/// A unit of data within a tag.
pub struct Frame {
	/// Uniquely identifies the frame type.
	pub name4: String,
	/// Includes 10-byte frame header.
	pub original_size: u32,
	/// Almost always zero.
	pub flags: (u8, u8),
	/// Polymorphic data.
	pub data: FrameData,
}

impl Frame {
	pub fn parse(src: &[u8]) -> w::AnyResult<Self> {
		let mut src = src;

		// Parse the 10-byte frame header.
		let name4 = str_engine::from_ascii(&src[0..4]);
		let original_size = u32::from_be_bytes(src[4..8].try_into()?) + 10; // also count 10-byte tag header
		let flags = (src[8], src[9]);

		// Skip frame header, truncate to frame size.
		src = &src[10..original_size as usize];

		// Parse the frame contents.
		let data = FrameData::parse(&name4, src)?;

		Ok(Self { name4, original_size, flags, data })
	}
}
