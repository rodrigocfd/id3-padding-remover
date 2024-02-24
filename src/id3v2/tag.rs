use winsafe::{self as w};

use super::consts::Field;
use super::frame::Frame;
use super::str_engine;
use super::synch_safe;

/// Metadata of a single MP3 file.
#[derive(Default)]
pub struct Tag {
	declared_size: u32,
	mp3_offset: u32,
	padding: u32,
	frames: Vec<Frame>,
}

impl Tag {
	/// Reads the tag from an MP3 file.
	#[must_use]
	pub fn read_from_file(mp3_path: &str) -> w::AnyResult<Self> {
		let fin = w::FileMapped::open(mp3_path, w::FileAccess::ExistingReadOnly)?;
		Self::parse(fin.as_slice())
	}

	/// Parses the tag from a binary blob.
	#[must_use]
	pub fn parse(src: &[u8]) -> w::AnyResult<Self> {
		let (declared_size, mp3_offset) = Self::parse_header(src)?;
		if declared_size == 0 && mp3_offset == 0 {
			return Ok(Self::default()); // file has no tag
		}

		let (frames, padding) = Self::parse_frames(&src[10..declared_size as _])?;
		Ok(Self { declared_size, mp3_offset, padding, frames })
	}

	/// Returns declared size and MP3 offset.
	#[must_use]
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

	/// Returns the frames and the padding.
	#[must_use]
	fn parse_frames(src: &[u8]) -> w::AnyResult<(Vec<Frame>, u32)> {
		let mut src = src;
		let mut frames = Vec::with_capacity(10); // arbitrary
		let mut padding = 0;

		loop {
			if src.is_empty() { // end of tag, no padding found
				break;
			} else if src.iter().all(|b| *b == 0x00) { // we entered a padding region after all frames
				padding = src.len() as _;
				break;
			}

			let new_frame = Frame::parse(src)?;
			if new_frame.original_size() > src.len() as _ { // means the size was serialized with error
				return Err(format!(
					"Frame size is greater than available size: {} vs {}.",
					new_frame.original_size(), src.len(),
				).into());
			}

			src = &src[new_frame.original_size() as _..];
			frames.push(new_frame); // add the frame to our collection
		}

		Ok((frames, padding))
	}

	/// Returns the given known field as a single string, or an empty string if
	/// the field is absent.
	#[must_use]
	pub fn field(&self, f: Field) -> String {
		self.frames.iter()
			.find(|frame| frame.name4() == f.name4())
			.map(|frame| frame.data().to_string())
			.unwrap_or_default()
	}

	/// Serializes the tag into a `Vec<u8>`.
	#[must_use]
	pub fn serialize(&self) -> Vec<u8> {
		let serialized_frames = self.frames.iter()
			.flat_map(|frame| frame.serialize())
			.collect::<Vec<_>>();
		let synch_safe_data_size = synch_safe::encode(serialized_frames.len() as _); // won't count 10-byte header

		str_engine::to_ascii("ID3").into_iter() // magic bytes
			.chain([0x03, 0x00].into_iter()) // tag version 2.3.0
			.chain([0x00].into_iter()) // flags
			.chain(synch_safe_data_size.to_be_bytes()) // data size is the last part of the 10-byte header
			.chain(serialized_frames.into_iter())
			.collect()
	}

	#[must_use]
	pub const fn declared_size(&self) -> u32 {
		self.declared_size
	}

	#[must_use]
	pub const fn mp3_offset(&self) -> u32 {
		self.mp3_offset
	}

	#[must_use]
	pub const fn padding(&self) -> u32 {
		self.padding
	}

	#[must_use]
	pub const fn frames(&self) -> &Vec<Frame> {
		&self.frames
	}
}
