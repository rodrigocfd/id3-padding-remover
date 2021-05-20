use std::convert::TryInto;
use std::error::Error;
use winsafe as w;
use winsafe::co;

use super::Frame;
use super::util;

pub struct Tag {
	frames: Vec<Frame>,
}

impl Tag {
	pub fn read(file: &str) -> Result<Self, Box<dyn Error>> {
		let (hfile, _) = w::HFILE::CreateFile(file, co::GENERIC::READ,
			co::FILE_SHARE::READ, None, co::DISPOSITION::OPEN_EXISTING,
			co::FILE_ATTRIBUTE::NORMAL, None)?;
		let bytes = hfile.ReadFile(hfile.GetFileSizeEx()? as _, None)?;
		hfile.CloseHandle()?;
		Self::parse(&bytes)
	}

	pub fn parse(src: &[u8]) -> Result<Self, Box<dyn Error>> {
		let total_tag_size = Self::parse_header(src)?;
		println!("Total tag size: {}", total_tag_size);
		Err("NOT HERE YET".into())
	}

	pub fn frames(&self) -> &Vec<Frame> {
		&self.frames
	}

	fn parse_header(src: &[u8]) -> Result<u32, Box<dyn Error>> {
		// Check ID3 magic bytes.
		let magic_str = ['I' as u8, 'D' as u8, '3' as u8];
		if src[..3] != magic_str {
			return Err("No ID3 tag found.".into());
		}

		// Validate tag version 2.3.0.
		if src[3..5] != [3, 0] { // the first "2" is not stored in the tag
			return Err(
				format!("Tag version 2.{}.{} is not supported, only 2.3.0.",
					src[3], src[4]).into(),
			);
		}

		// Validate unsupported flags.
		if (src[5] & 0b1000_0000) != 0 {
			return Err("Tag is unsynchronised, not supported.".into());
		} else if (src[5] & 0b0100_0000) != 0 {
			return Err("Tag extended header not supported.".into());
		}

		// Read total tag size.
		let total_tag_size = util::SynchSafeDecode(
			u32::from_be_bytes(src[6..10].try_into()?),
		) + 10; // also count 10-byte tag header

		Ok(total_tag_size)
	}
}
