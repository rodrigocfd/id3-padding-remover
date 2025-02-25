use winsafe::{self as w};

use super::body::Body;
use super::frame::Frame;
use super::str_engine;
use super::synch_safe;

/// Metadata of a single MP3 file.
#[derive(Default)]
pub struct Tag {
	mp3_offset: u32,
	padding: u32,
	frames: Vec<Frame>,
}

impl std::fmt::Display for Tag {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		write!(f, "Off: {}, pad: {}\n{}",
			self.mp3_offset,
			self.padding,
			self.frames.iter()
				.map(|f| f.to_string())
				.collect::<Vec<_>>()
				.join("\n"),
		)
	}
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
			Ok(Self::default()) // file has no tag
		} else {
			let (frames, padding) = Self::parse_frames(&src[10..declared_size as _])?;
			Ok(Self { mp3_offset, padding, frames })
		}
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

			let (new_frame, original_size) = Frame::parse(src)?;
			if original_size > src.len() as _ { // means the size was serialized with error
				return Err(format!(
					"Frame size is greater than available size: {} vs {}.",
					original_size, src.len(),
				).into());
			}

			src = &src[original_size as _..];
			frames.push(new_frame); // add the frame to our collection
		}

		Ok((frames, padding))
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

	/// Saves the tag to an MP3 file. If there are no frames, the tag will be
	/// entirely removed from the file.
	pub fn save_to_file(&self, mp3_path: &str) -> w::AnyResult<()> {
		let fout = w::File::open(mp3_path, w::FileAccess::ExistingRW)?;
		let current_contents = fout.read_all()?; // read the whole MP3 into a buffer
		let current_tag = Self::parse(&current_contents)?; // parse tag currently saved in the MP3 file

		if self.frames.is_empty() {
			fout.erase_and_write(
				&current_contents[current_tag.mp3_offset as _..], // no tag will be written
			)?;
		} else {
			fout.erase_and_write(
				&self.serialize().into_iter()
					.chain(
						current_contents[current_tag.mp3_offset as _..].iter()
							.map(|b| *b),
					)
					.collect::<Vec<_>>(),
			)?;
		}

		Ok(())
	}

	#[must_use]
	pub const fn padding(&self) -> u32 {
		self.padding
	}

	#[must_use]
	pub const fn frames(&self) -> &Vec<Frame> {
		&self.frames
	}

	#[must_use]
	pub const fn frames_mut(&mut self) -> &mut Vec<Frame> {
		&mut self.frames
	}

	#[must_use]
	pub fn frame(&self, name4: &str) -> Option<&Frame> {
		self.frames.iter()
			.find(|frame| frame.name4() == name4)
	}

	/// Any ReplayGain frame present?
	#[must_use]
	pub fn has_replay_gain(&self) -> bool {
		self.frames.iter()
			.find(|frame| {
				if frame.name4() == "TXXX" {
					if let Body::UserText(ut) = frame.body() {
						if ut.descr.starts_with("replaygain") {
							return true;
						}
					}
				}
				false
			})
			.is_some()
	}

	/// Tries to set the frame value as a simple text, returning an error if not
	/// possible. If frame does not exist, creates it.
	pub fn set_frame_str(&mut self, name4: &str, text: &str) -> w::AnyResult<()> {
		if text.is_empty() {
			if let Some(idx) = self.frames.iter().position(|frame| frame.name4() == name4) {
				self.frames.remove(idx); // empty string will remove frame
			}
		} else { // text is not empty
			match self.frames.iter_mut().find(|frame| frame.name4() == name4) {
				Some(frame) => frame.set_string(text)?, // field exists, update
				None => self.frames.push(Frame::new_from_string(name4, text)?), // create simple text frame
			}
		}
		Ok(())
	}
}
