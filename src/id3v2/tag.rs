use winsafe::{self as w};

use super::frame::Frame;
use super::synch_safe;

/// Metadata of a single MP3 file.
#[derive(Default)]
pub struct Tag {
	pub declared_size: u32,
	pub mp3_offset: u32,
	pub padding: u32,
	pub frames: Vec<Frame>,
}

impl Tag {
	/// Returns declared size and MP3 offset.
	fn parse_header(src: &[u8]) -> w::AnyResult<(u32, u32)> {
		// Find MP3 offset.
		let mp3_offset = match src.windows(2)
			.position(|bb| bb == &[0xff, 0xfb]) // https://stackoverflow.com/a/7302482/6923555
			.map(|idx| idx as u32)
		{
			Some(idx) => idx,
			None => return Err(format!("No MP3 signature found.").into()),
		};

		// Check ID3 magic bytes.
		if &src[..3] != &['I' as u8, 'D' as u8, '3' as u8] {
			return Ok((0, mp3_offset)); // MP3 file has no tag
		}

		// Validate tag version 2.3.0.
		if &src[3..5] != &[3, 0] { // the first "2" is not stored in the tag
			return Err(format!(
				"Tag version 2.{}.{} is not supported, only 2.3.0.",
				src[3], src[4],
			).into());
		}

		// Validate unsupported flags.
		if src[5] & 0b1000_0000 != 0 {
			return Err("Unsynchronised tag not supported.".into());
		} else if src[5] & 0b0100_0000 != 0 {
			return Err("Tag extended header not supported.".into());
		}

		// Read declared tag size.
		let declared_size = synch_safe::decode(
			u32::from_be_bytes(src[6..10].try_into()?)) + 10; // also count 10-byte tag header

		if declared_size > mp3_offset {
			return Err(format!(
				"Declared size is greater than MP3 offset: {} vs {}.",
				declared_size, mp3_offset,
			).into());
		}

		Ok((declared_size, mp3_offset))
	}
}
