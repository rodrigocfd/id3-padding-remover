use winsafe::{self as w};

use super::body::Body;
use super::frame::Frame;
use super::str_engine;
use super::synch_safe;

/// Metadata of a single MP3 file.
#[derive(Default, Clone)]
pub struct Tag {
	mp3_offset: usize,
	padding: usize,
	frames: Vec<Frame>,
}

impl std::fmt::Display for Tag {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		write!(
			f,
			"Off: {}, pad: {}\n{}",
			self.mp3_offset,
			self.padding,
			self.frames
				.iter()
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
		let (_, mp3_offset) = Self::parse_header(src)?; // discard declared size
		if mp3_offset == 0 {
			Ok(Self::default()) // MP3 file has no tag
		} else {
			let (frames, padding) = Self::parse_frames(&src[10..mp3_offset])?;
			Ok(Self { mp3_offset, padding, frames })
		}
	}

	/// Returns declared size and MP3 offset.
	#[must_use]
	fn parse_header(mut src: &[u8]) -> w::AnyResult<(u32, usize)> {
		// Find MP3 offset.
		let mp3_offset = match src
			.windows(2)
			.position(|bb| bb == &[0xff, 0xfb]) // https://stackoverflow.com/a/7302482/6923555
		{
			Some(idx) => match idx {
				0 => return Ok((0, 0)), // MP3 file has no tag
				idx => idx,
			},
			None => return Err(format!("No MP3 signature found.").into()),
		};
		src = &src[0..mp3_offset]; // limit our range

		// Check ID3 magic bytes.
		if &src[..3] != &str_engine::to_ascii("ID3") {
			return Ok((0, mp3_offset)); // MP3 file has no tag
		}

		// Validate tag version 2.3.0.
		// The first "2" is not stored in the tag.
		if &src[3..5] != &[3, 0] {
			return Err(format!(
				"Tag version 2.{}.{} is not supported, only 2.3.0.",
				src[3], src[4],
			)
			.into());
		}

		// Validate unsupported flags.
		if src[5] & 0b1000_0000 != 0 {
			return Err("Unsynchronised tag not supported.".into());
		} else if src[5] & 0b0100_0000 != 0 {
			return Err("Tag extended header not supported.".into());
		}

		// Read declared tag size; also count 10-byte tag header.
		let declared_size = synch_safe::decode(u32::from_be_bytes(src[6..10].try_into()?)) + 10;

		Ok((declared_size, mp3_offset))
	}

	/// Returns the frames and the padding.
	#[must_use]
	fn parse_frames(mut src: &[u8]) -> w::AnyResult<(Vec<Frame>, usize)> {
		let mut frames = Vec::with_capacity(10); // arbitrary
		let mut padding = 0usize;

		loop {
			if src.is_empty() {
				break; // end of tag, no padding found
			} else if src.len() < 10 {
				// We cant' have a frame with less than 10 bytes, which is the header size.
				padding = src.len();
				break;
			} else if src[0..4].iter().all(|b| *b == 0x00) {
				// The first 4 bytes should contain the 4-char frame name.
				// If they're all zero, it means we entered a padding region after all frames.
				padding = src.len();
				break;
			}

			let (new_frame, declared_size) = Frame::parse(src)?;
			if declared_size > src.len() {
				// Means the size was serialized with error.
				return Err(format!(
					"Frame size is greater than available size: {} vs {}.",
					declared_size,
					src.len(),
				)
				.into());
			}

			src = &src[declared_size..];
			frames.push(new_frame); // add the frame to our collection
		}

		Ok((frames, padding))
	}

	/// Serializes the tag into a `Vec<u8>`.
	#[must_use]
	pub fn serialize(&self) -> Vec<u8> {
		let serialized_frames = self
			.frames
			.iter()
			.flat_map(|frame| frame.serialize())
			.collect::<Vec<_>>();
		let synch_safe_data_size = synch_safe::encode(serialized_frames.len() as _); // won't count 10-byte header

		str_engine::to_ascii("ID3")
			.into_iter() // magic bytes
			.chain([0x03, 0x00].into_iter()) // tag version 2.3.0
			.chain([0x00].into_iter()) // flags
			.chain(synch_safe_data_size.to_be_bytes()) // data size is the last part of the 10-byte header
			.chain(serialized_frames.into_iter()) // then all the serialized frames
			.collect()
	}

	/// Saves the tag to an MP3 file. If there are no frames, the tag will be
	/// entirely removed from the file.
	///
	/// No padding will be written, and the `padding` field will be set to zero.
	pub fn save_to_file(&mut self, mp3_path: &str) -> w::AnyResult<()> {
		let fout = w::File::open(mp3_path, w::FileAccess::ExistingRW)?;
		let current_contents = fout.read_all()?; // read the whole MP3 into a buffer
		let (_, mp3_offset) = Self::parse_header(&current_contents)?;

		if self.frames.is_empty() {
			fout.erase_and_write(
				&current_contents[mp3_offset..], // no tag will be written
			)?;
		} else {
			fout.erase_and_write(
				&self
					.serialize()
					.into_iter()
					.chain(current_contents[mp3_offset..].iter().map(|b| *b)) // MP3 data
					.collect::<Vec<_>>(),
			)?;
		}
		self.padding = 0;
		Ok(())
	}

	#[must_use]
	pub const fn padding(&self) -> usize {
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
	pub fn frame_by_name4(&self, name4: &str) -> Option<&Frame> {
		self.frames.iter().find(|frame| frame.name4() == name4)
	}

	/// If the given frame does exist, tries to set the editable string on it.
	///
	/// If the frame doesn't exist, creates a new one with the editable string
	/// on it.
	///
	/// An empty string will remove the frame, if any.
	pub fn set_editable_string(&mut self, name4: &str, val: &str) -> w::AnyResult<()> {
		let val = val.trim();
		if val.is_empty() {
			self.frames.retain(|frame| frame.name4() != name4); // remove empty new values
		} else {
			match self.frames.iter_mut().find(|frame| frame.name4() == name4) {
				Some(frame) => frame.set_editable_string(val)?, // frame already exists, set new value
				None => {
					let new_frame = Frame::new_from_editable_string(name4, val)?;
					self.frames.push(new_frame); // frame doesn't exist, push new
				},
			}
		}
		Ok(())
	}

	/// Any ReplayGain frame present?
	#[must_use]
	pub fn has_replay_gain(&self) -> bool {
		self.frames
			.iter()
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
}

/// Returns true if the given frame is equal across all given tags.
pub fn equal_frame_across_all_tags(name4: &str, tags: &[Tag]) -> bool {
	if tags.is_empty() {
		return false; // nothing to do
	}

	let frame0 = match tags[0].frame_by_name4(name4) {
		Some(f) => f,
		None => return false, // the 1st tag doesn't have this frame
	};

	tags.iter().skip(1).all(|tag| {
		match tag.frame_by_name4(name4) {
			Some(f) => f == frame0, // frame present, check equality
			None => false,          // this tag doesn't have this frame
		}
	})
}
