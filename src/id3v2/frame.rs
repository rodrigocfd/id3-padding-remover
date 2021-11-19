use std::convert::TryInto;
use winsafe as w;

use super::FrameComment;
use super::util;

/// The data contained in a frame, which can be of various types.
#[derive(PartialEq, Eq)]
pub enum FrameData {
	Text(String),
	MultiText(Vec<String>),
	Comment(FrameComment),
	Binary(Vec<u8>),
}

/// A unit of data in an MP3 tag.
#[derive(PartialEq, Eq)]
pub struct Frame {
	name4:         String,
	original_size: usize,
	flags:         [u8; 2],
	data:          FrameData,
}

impl Frame {
	pub fn name4(&self) -> &str { &self.name4 }
	pub const fn flags(&self) -> &[u8; 2] { &self.flags }
	pub const fn original_size(&self) -> usize { self.original_size }
	pub const fn data(&self) -> &FrameData { &self.data }
	pub fn data_mut(&mut self) -> &mut FrameData { &mut self.data }

	pub fn new(name4: &str, data: FrameData) -> Self {
		Self {
			name4: name4.to_owned(),
			original_size: 0,
			flags: [0x00, 0x00],
			data,
		}
	}

	pub fn parse(mut src: &[u8]) -> w::ErrResult<Self> {
		let name4 = std::str::from_utf8(&src[0..4])?.to_string();
		let original_size = u32::from_be_bytes(src[4..8].try_into()?) as usize + 10; // also count 10-byte tag header
		let flags: [u8; 2] = src[8..10].try_into()?;

		src = &src[10..original_size]; // skip frame header, truncate to frame size

		let data = if name4 == "COMM" {
			FrameData::Comment(FrameComment::parse(src)?)
		} else if name4.chars().nth(0).unwrap() == 'T' { // text frame
			let texts = util::parse_strs::any(src)?;
			match texts.len() {
				0 => return Err(format!("Frame {} contains no texts.", name4).into()),
				1 => FrameData::Text(texts[0].clone()),
				_ => FrameData::MultiText(texts),
			}
		} else { // anything else will be treated as raw binary
			FrameData::Binary(src.to_vec())
		};

		Ok(Self { name4, original_size, flags, data })
	}

	pub fn serialize(&self) -> w::ErrResult<Vec<u8>> {
		if self.name4.chars().count() != 4 {
			return Err(format!("Frame name length is not 4 [{}]", self.name4).into());
		}

		let frame_data = match &self.data {
			FrameData::Text(text) => util::SerializedStrs::new(&[&text]).collect(),
			FrameData::MultiText(texts) => util::SerializedStrs::new(&texts).collect(),
			FrameData::Comment(comm) => comm.serialize_data(),
			FrameData::Binary(bin) => bin.clone(),
		};

		let mut buf = Vec::<u8>::with_capacity(frame_data.len() + 10);
		buf.extend(
			self.name4.chars()
				.enumerate()
				.map(|(_, ch)| ch as u8),
		);
		buf.extend_from_slice(&(frame_data.len() as u32).to_be_bytes()); // size not counting 10-byte header
		buf.extend_from_slice(&self.flags); // flags
		buf.extend_from_slice(&frame_data);

		Ok(buf)
	}
}
