use std::convert::TryInto;
use std::error::Error;
use winsafe as w;
use winsafe::co;

use super::Frame;
use super::util;

/// The MP3 file metadata.
pub struct Tag {
	frames:           Vec<Frame>,
	original_size:    usize,
	original_padding: usize,
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

	pub fn parse(mut src: &[u8]) -> Result<Self, Box<dyn Error>> {
		let original_size = Self::parse_header(src)?;
		src = &src[10..original_size]; // skip 10-byte tag header; truncate to tag bounds

		let (frames, original_padding) = Self::parse_all_frames(src)?;

		Ok(Self { frames, original_size, original_padding })
	}

	pub fn original_size(&self) -> usize {
		self.original_size
	}

	pub fn original_padding(&self) -> usize {
		self.original_padding
	}

	pub fn frames(&self) -> &Vec<Frame> {
		&self.frames
	}

	pub fn frames_mut(&mut self) -> &mut Vec<Frame> {
		&mut self.frames
	}

	fn parse_header(src: &[u8]) -> Result<usize, Box<dyn Error>> {
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
		let total_tag_size = util::synch_safe_decode(
			u32::from_be_bytes(src[6..10].try_into()?),
		) as usize + 10; // also count 10-byte tag header

		Ok(total_tag_size)
	}

	fn parse_all_frames(mut src: &[u8]) -> Result<(Vec<Frame>, usize), Box<dyn Error>> {
		let mut frames = Vec::new();
		let mut original_padding = 0;

		loop {
			if src.is_empty() { // end of tag, no padding found
				break;
			} else if util::is_all_zero(src) { // we entered a padding region after all frames
				original_padding = src.len();
				break;
			}

			let new_frame = Frame::parse(src)?;
			frames.push(new_frame);
			let new_frame_ref = frames.last().unwrap();

			if new_frame_ref.original_size() > src.len() { // means the tag was serialized with error
				return Err(
					format!(
						"Frame size is greater than declared tag size: {} vs {}.",
						new_frame_ref.original_size(),
						src.len(),
					).into(),
				);
			}

			src = &src[new_frame_ref.original_size()..]; // now starts at 1st byte of next frame
		}

		Ok((frames, original_padding))
	}
}
