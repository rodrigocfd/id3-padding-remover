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
		match Self::parse_header(src)? {
			Some(_declared_size) => {
				let (frames, mp3_offset, padding) = Self::parse_frames(&src[10..])?;
				Ok(Self { mp3_offset, padding, frames })
			},
			None => Ok(Self::default()), // MP3 file has no ID3v2 tag
		}
	}

	/// If an ID3v2 tag is present, return its declared size, including the
	/// 10-byte header.
	#[must_use]
	fn parse_header(src: &[u8]) -> w::AnyResult<Option<u32>> {
		// Check ID3 magic bytes.
		if &src[..3] != &str_engine::to_ascii("ID3") {
			return Ok(None); // MP3 file has no tag
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
		Ok(Some(declared_size))
	}

	/// Returns the frames, MP3 offset and padding size.
	#[must_use]
	fn parse_frames(mut src: &[u8]) -> w::AnyResult<(Vec<Frame>, usize, usize)> {
		let mut frames = Vec::with_capacity(10); // arbitrary
		let mut offset = 10usize; // start at 10 because src already skipped 10-byte header

		// Two known magic byte sequences that identify the beginning of the MP3.
		// https://stackoverflow.com/a/7302482/6923555
		// https://en.wikipedia.org/wiki/List_of_file_signatures
		// https://github.com/sindresorhus/file-type/issues/75#issuecomment-320650344
		const MP3_MAGIC: [[u8; 2]; 5] =
			[[0xff, 0xfb], [0xff, 0xfb], [0xff, 0xf2], [0xff, 0xfa], [0xff, 0xf3]];

		loop {
			if MP3_MAGIC.iter().any(|magic| magic == &src[0..2]) {
				// We found the beginning of the MP3 file, no padding.
				return Ok((frames, offset, 0));
			} else if src[0] == 0x0000 {
				// We entered a padding region after all frames.
				match src
					.windows(2)
					.position(|by| MP3_MAGIC.iter().any(|magic| magic == by))
				{
					Some(idx_mp3_start) => {
						return Ok((frames, offset + idx_mp3_start, idx_mp3_start));
					},
					None => return Err("MP3 offset not found.".into()),
				}
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

			offset += declared_size;
			src = &src[declared_size..];
			frames.push(new_frame); // add the frame to our collection
		}
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
		let old_tag = Self::parse(&current_contents)?; // so we can extract MP3 offset

		if self.frames.is_empty() {
			fout.erase_and_write(
				&current_contents[old_tag.mp3_offset..], // no tag will be written
			)?;
		} else {
			fout.erase_and_write(
				&self
					.serialize()
					.into_iter()
					.chain(current_contents[old_tag.mp3_offset..].iter().map(|b| *b)) // MP3 data
					.collect::<Vec<_>>(),
			)?;
		}
		self.padding = 0; // because we write no padding
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
