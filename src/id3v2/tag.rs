use std::convert::TryInto;
use winsafe::{self as w, ErrResult};

use super::{Frame, FrameData, TextField};
use super::util;

/// The MP3 file metadata.
pub struct Tag {
	frames:           Vec<Frame>,
	original_size:    usize,
	original_padding: usize,
}

impl Tag {
	pub fn read(file: &str) -> ErrResult<Self> {
		let mapped_file = w::FileMapped::open(file, w::FileAccess::ExistingReadOnly)?;
		Self::parse(mapped_file.as_slice())
	}

	pub fn parse(mut src: &[u8]) -> ErrResult<Self> {
		let original_size = Self::_parse_header(src)?;
		src = &src[10..original_size]; // skip 10-byte tag header; truncate to tag bounds

		let (frames, original_padding) = Self::_parse_all_frames(src)?;

		Ok(Self { frames, original_size, original_padding })
	}

	/// Returns the original tag size, including 10-byte header and padding.
	pub const fn original_size(&self) -> usize {
		self.original_size
	}

	/// Returns the original padding size.
	pub const fn original_padding(&self) -> usize {
		self.original_padding
	}

	/// Replaces the tag in the given MP3 file with this one.
	pub fn write(&self, file: &str) -> ErrResult<()> {
		let blob_new = self._serialize();
		let mut mapped_file = w::FileMapped::open(file, w::FileAccess::ExistingReadWrite)?;
		let file_size_old = mapped_file.size();
		let tag_old = Self::parse(mapped_file.as_slice())?;

		// Calculate size difference between new/old tags.
		let diff = blob_new.len() as isize - tag_old.original_size() as isize;

		if diff > 0 { // new tag is larger, we need to make room
			mapped_file.resize(mapped_file.size() + diff as usize)?;
		}

		// Move the MP3 data block inside the file.
		let hot_slice = mapped_file.as_mut_slice();
		hot_slice.copy_within(
			tag_old.original_size()..file_size_old,
			(tag_old.original_size() as isize + diff) as _,
		);

		// Copy the new tag blob into the file room, no padding.
		hot_slice[0..blob_new.len()].copy_from_slice(&blob_new);

		if diff < 0 { // new tag is shorter, shrink
			mapped_file.resize((mapped_file.size() as isize + diff) as _)?;
		}

		Ok(())
	}

	fn _parse_header(src: &[u8]) -> ErrResult<usize> {
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
			u32::from_be_bytes(src[6..10].try_into()?), // https://stackoverflow.com/a/50080940/6923555
		) as usize + 10; // also count 10-byte tag header

		Ok(total_tag_size)
	}

	fn _parse_all_frames(mut src: &[u8]) -> ErrResult<(Vec<Frame>, usize)> {
		let mut frames = Vec::default();
		let mut original_padding = 0;

		loop {
			if src.is_empty() { // end of tag, no padding found
				break;
			} else if src.iter().find(|b| **b != 0x00).is_none() {
				// If the rest of the blob contains only zeros,
				// we entered a padding region.
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

	fn _serialize(&self) -> Vec<u8> {
		let mut frames_buf: Vec<u8> = Vec::with_capacity(100); // arbitrary
		for f in self.frames.iter() {
			frames_buf.extend_from_slice(&f.serialize());
		}

		let mut buf: Vec<u8> = Vec::with_capacity(10 + frames_buf.len());
		buf.extend(&['I' as u8, 'D' as u8, '3' as u8]); // magic bytes
		buf.extend_from_slice(&[0x03, 0x00]); // tag version
		buf.push(0x00); // flags
		buf.extend_from_slice(&util::synch_safe_encode(frames_buf.len() as _).to_be_bytes()); // tag size, minus header
		buf.extend_from_slice(&frames_buf);

		buf
	}

	/// Returns a reference to the frame array.
	pub const fn frames(&self) -> &Vec<Frame> {
		&self.frames
	}

	/// Returns a mutable reference to the frame array.
	pub fn frames_mut(&mut self) -> &mut Vec<Frame> {
		&mut self.frames
	}

	pub fn text_field(&self, text_field: TextField) -> Option<ErrResult<&str>> {
		let (name4, fancy_name) = text_field.names();
		self.frames.iter()
			.find(|f| f.name4() == name4)
			.map(|f| match f.data() {
				FrameData::Comment(comm) => Ok(comm.text.as_ref()), // ignore comment lang
				FrameData::Text(text) => Ok(text.as_str()),
				_ => Err(format!("{} has wrong frame type.", fancy_name).into()),
			})
	}

	pub fn text_field_mut(&mut self, text_field: TextField) -> Option<ErrResult<&mut String>> {
		let (name4, fancy_name) = text_field.names();
		self.frames.iter_mut()
			.find(|f| f.name4() == name4)
			.map(|f| match f.data_mut() {
				FrameData::Comment(comm) => Ok(&mut comm.text), // ignore comment lang
				FrameData::Text(text) => Ok(text),
				_ => Err(format!("{} has wrong frame type.", fancy_name).into()),
			})
	}

	/// Tells whether `TextField` is the same among all given tags.
	pub fn is_uniform_text_field<'a>(
		tags: &'a [(&'a String, &'a Self)],
		text_field: TextField) -> ErrResult<Option<&'a str>>
	{
		if tags.len() == 0 { // no tags to look at
			return Ok(None);
		} else if tags.len() == 1 { // 1 single tag
			return if let Some(field) = tags[0].1.text_field(text_field) {
				Ok(Some(field?)) // good value
			} else {
				return Ok(None) // frame does not exist in tag
			}
		}

		let (name4, _) = text_field.names();

		let mut tags_iter = tags.iter().map(|(_, tag)| tag);
		let first_tag = tags_iter.next().unwrap(); // extract first tag from iterator
		let first_frame = match first_tag.frames.iter().find(|f| f.name4() == name4) {
			Some(frame) => frame,
			None => return Ok(None), // frame absence in first tag
		};

		for other_tag in tags_iter {
			let other_frame = match other_tag.frames.iter().find(|f| f.name4() == name4) {
				Some(frame) => frame,
				None => return Ok(None), // frame absence in subsequent tag
			};

			if first_frame != other_frame {
				return Ok(None); // this frame is different from the first one
			}
		}

		if let Some(field) = tags[0].1.text_field(text_field) { // yes, all frames are equal
			Ok(Some(field?)) // good value
		} else {
			return Ok(None) // frame does not exist in tag
		}
	}
}
