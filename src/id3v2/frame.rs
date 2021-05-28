use std::convert::TryInto;
use std::error::Error;

use super::FrameComment;
use super::util;

/// The data contained in a frame, which can be of various types.
pub enum FrameData {
	Text(String),
	MultiText(Vec<String>),
	Comment(FrameComment),
	Binary(Vec<u8>),
}

/// A unit of data in an MP3 tag.
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
		} else if name4.chars().nth(0).unwrap() == 'T' { // text frame
			let texts = util::parse_any_strings(src)?;
			match texts.len() {
				0 => return Err(format!("Frame {} contains no texts.", name4).into()),
				1 => FrameData::Text(texts[0].clone()),
				_ => FrameData::MultiText(texts),
			}
		} else { // anything else will be treated as raw binary
			FrameData::Binary(src.to_vec())
		};

		Ok(Self { name4, original_size, data })
	}

	/// Returns the 4-character frame ID.
	pub fn name4(&self) -> &str {
		&self.name4
	}

	/// Returns the original frame size, including 10-byte header.
	pub fn original_size(&self) -> usize {
		self.original_size
	}

	/// Returns the data of the frame, which can be of various types.
	pub fn data(&self) -> &FrameData {
		&self.data
	}

	/// Returns the mutable data of the frame, which can be of various types.
	pub fn data_mut(&mut self) -> &mut FrameData {
		&mut self.data
	}

	/// Serializes the frame into bytes.
	pub fn serialize(&self) -> Vec<u8> {
		let frame_data = match &self.data {
			FrameData::Text(text) => util::SerializedStrs::new(&[&text]).collect(),
			FrameData::MultiText(texts) => util::SerializedStrs::new(&texts).collect(),
			FrameData::Comment(comm) => comm.serialize_data(),
			FrameData::Binary(bin) => bin.clone(),
		};

		let mut buf: Vec<u8> = Vec::with_capacity(frame_data.len() + 10);
		buf.extend(
			self.name4.chars().enumerate()
				.map(|(_, ch)| ch as u8),
		);
		buf.extend_from_slice(&(frame_data.len() as u32).to_be_bytes()); // size not counting 10-byte header
		buf.extend_from_slice(&[0x00, 0x00]); // flags
		buf.extend_from_slice(&frame_data);

		buf
	}
}
