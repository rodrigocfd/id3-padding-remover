use std::convert::TryInto;
use std::error::Error;

use super::FrameComment;
use super::FrameData;
use super::util;

pub struct Frame {
	name4:         String,
	original_size: usize,
	data:          FrameData,
}

impl Frame {
	pub fn parse(mut src: &[u8]) -> Result<Self, Box<dyn Error>> {
		let name4 = std::str::from_utf8(&src[0..4])?.to_string();
		let original_size = u32::from_be_bytes(src[4..8].try_into()?) as usize + 10; // also count 10-byte tag header

		src = &src[10..original_size]; // skip frame header, truncate to frame size

		let data = if name4 == "COMM" {
			FrameData::Comment(FrameComment::parse(src)?)
		} else if name4.chars().nth(0).unwrap() == 'T' {
			let texts = util::parse_any_strings(src)?;
			match texts.len() {
				0 => return Err(format!("Frame {} contains no texts.", name4).into()),
				1 => FrameData::Text(texts[0].clone()),
				_ => FrameData::MultiText(texts),
			}
		} else {
			FrameData::Binary(src.to_vec())
		};

		Ok(Self { name4, original_size, data })
	}

	pub fn name4(&self) -> &str {
		&self.name4
	}

	pub fn original_size(&self) -> usize {
		self.original_size
	}

	pub fn data(&self) -> &FrameData {
		&self.data
	}
}
