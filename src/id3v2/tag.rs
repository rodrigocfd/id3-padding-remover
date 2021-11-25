use std::convert::TryInto;
use winsafe as w;

use super::{Frame, FrameComment, FrameData, FieldName};
use super::util;

/// The MP3 file metadata.
pub struct Tag {
	declared_size: usize,
	mp3_offset:    usize,
	padding:       usize,
	frames:        Vec<Frame>,
}

impl Tag {
	pub const fn declared_size(&self) -> usize { self.declared_size }
	pub const fn mp3_offset(&self) -> usize { self.mp3_offset }
	pub const fn padding(&self) -> usize { self.padding }
	pub fn is_empty(&self) -> bool { self.frames.is_empty() }
	pub const fn frames(&self) -> &Vec<Frame> { &self.frames }
	pub fn frames_mut(&mut self) -> &mut Vec<Frame> { &mut self.frames }

	pub fn new_empty() -> Self {
		Self {
			declared_size: 0,
			mp3_offset: 0,
			padding: 0,
			frames: Vec::default(),
		}
	}

	pub fn read(file: &str) -> w::ErrResult<Self> {
		let mapped_file = w::FileMapped::open(file, w::FileAccess::ExistingReadOnly)?;
		Self::parse(mapped_file.as_slice())
	}

	pub fn parse(src: &[u8]) -> w::ErrResult<Self> {
		let (declared_size, mp3_offset) = Self::_parse_header(src)?;

		if declared_size == 0 && mp3_offset == 0 {
			return Ok(Self::new_empty()); // file has no tag
		}

		let src = &src[10..declared_size]; // skip 10-byte tag header; truncate to tag bounds
		let (frames, padding) = Self::_parse_all_frames(src)?;

		Ok(Self { declared_size, mp3_offset, padding, frames })
	}

	fn _parse_header(src: &[u8]) -> w::ErrResult<(usize, usize)> {
		let mp3_offset = match src.windows(2)
			.position(|by| by == [0xff, 0xfb]) // https://stackoverflow.com/a/7302482/6923555
		{
			None => return Err("No MP3 signature found.".into()),
			Some(mp3_off) => mp3_off,
		};

		// Check ID3 magic bytes.
		let magic_str = "ID3".as_bytes();
		if &src[..3] != magic_str {
			return Ok((0, 0)); // no tag found
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

		// Read declared tag size.
		let declared_tag_size = util::synch_safe::decode(
			u32::from_be_bytes(src[6..10].try_into()?), // https://stackoverflow.com/a/50080940/6923555
		) as usize + 10; // also count 10-byte tag header

		if declared_tag_size > mp3_offset {
			return Err(
				format!("Declared size is greater than MP3 offset: {} vs {}.",
					declared_tag_size, mp3_offset).into(),
			);
		}

		Ok((declared_tag_size, mp3_offset))
	}

	fn _parse_all_frames(mut src: &[u8]) -> w::ErrResult<(Vec<Frame>, usize)> {
		let mut frames = Vec::<Frame>::with_capacity(10); // arbitrary
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
			if new_frame.original_size() > src.len() { // means the tag was serialized with error
				return Err(
					format!("Frame size is greater than declared tag size: {} vs {}.",
						new_frame.original_size(), src.len()).into(),
				);
			}

			src = &src[new_frame.original_size()..]; // now starts at 1st byte of next frame
			frames.push(new_frame);
		}

		Ok((frames, original_padding))
	}

	pub fn write(&self, file: &str) -> w::ErrResult<()> {
		let mut blob_new = Vec::<u8>::default(); // if tag is empty, this will actually remove any existing tag
		if !self.is_empty() {
			blob_new = self._serialize()?;
		}

		let mut fout = w::FileMapped::open(file, w::FileAccess::ExistingReadWrite)?;
		let fout_orig_size = fout.size();
		let current_tag = Self::parse(fout.as_slice())?;

		// Calculate size difference between new/old tags.
		let diff = blob_new.len() as isize - current_tag.mp3_offset() as isize;

		if diff > 0 { // new tag is larger, we need to make room
			fout.resize(fout.size() + diff as usize)?;
		}

		// Move the MP3 data block inside the file.
		let hot_slice = fout.as_mut_slice();
		hot_slice.copy_within(
			current_tag.mp3_offset()..fout_orig_size,
			(current_tag.mp3_offset() as isize + diff) as _,
		);

		// Copy the new tag blob into the file room, no padding.
		hot_slice[0..blob_new.len()].copy_from_slice(&blob_new);

		if diff < 0 { // new tag is shorter, shrink
			fout.resize((fout.size() as isize + diff) as _)?;
		}

		Ok(())
	}

	fn _serialize(&self) -> w::ErrResult<Vec<u8>> {
		let mut frames_buf = Vec::<u8>::with_capacity(100); // arbitrary
		for f in self.frames.iter() {
			frames_buf.extend_from_slice(&f.serialize()?);
		}

		let mut buf = Vec::<u8>::with_capacity(10 + frames_buf.len());
		buf.extend_from_slice("ID3".as_bytes()); // magic bytes
		buf.extend_from_slice(&[0x03, 0x00]); // tag version
		buf.push(0x00); // flags
		buf.extend_from_slice(&util::synch_safe::encode(frames_buf.len() as _).to_be_bytes()); // tag size, minus header
		buf.extend_from_slice(&frames_buf);

		Ok(buf)
	}

	pub fn text_by_field(&self,
		text_field: FieldName) -> w::ErrResult<Option<&str>>
	{
		let (name4, fancy_name) = text_field.names();
		self.frames.iter()
			.find(|f| f.name4() == name4)
			.map_or(Ok(None), // no such frame
				|f| match f.data() {
					FrameData::Comment(comm) => Ok(Some(&comm.text)), // ignore comment lang
					FrameData::Text(text) => Ok(Some(text)),
					_ => Err(format!("{} has wrong frame type.", fancy_name).into()),
				})
	}

	pub fn set_text_by_field(&mut self,
		text_field: FieldName,
		new_val: &str) -> w::ErrResult<()>
	{
		let (name4, fancy_name) = text_field.names();

		if new_val.is_empty() { // an empty string will delete the frame
			self.frames.retain(|f| f.name4() != name4);

		} else {
			if let Some(f) = self.frames.iter_mut()
				.find(|f| f.name4() == name4) // frame exists, update text
			{
				match f.data_mut() {
					FrameData::Comment(comm) => comm.set_text(new_val),
					FrameData::Text(text) => *text = new_val.to_owned(),
					_ => return Err(
						format!("Cannot set text on frame {} ({}).",
							name4, fancy_name).into(),
					),
				}

			} else { // no such frame yet, create new
				self.frames.push(
					Frame::new(name4, match text_field {
						FieldName::Comment => FrameData::Comment(FrameComment::new(None, None, new_val)),
						_ => FrameData::Text(new_val.to_owned()),
					}),
				);
			}
		}

		Ok(())
	}

	/// Tells whether the field value is the same among all given tags.
	pub fn same_field_value(
		tags: &Vec<&Self>,
		field_name: FieldName) -> w::ErrResult<Option<String>>
	{
		if tags.is_empty() { // no tags to look at
			return Ok(None);
		} else if tags.len() == 1 { // 1 single tag
			return Ok(
				tags[0]
					.text_by_field(field_name)?
					.map(|s| s.to_owned()),
			);
		}

		let mut tags_iter = tags.iter();
		let first_tag = tags_iter.next().unwrap(); // take first tag from iterator

		let first_val = match first_tag
			.text_by_field(field_name)?
			.map(|s| s.to_owned())
		{
			Some(val) => val,
			None => return Ok(None), // 1st tag doesn't have such field
		};

		let mut is_uniform = true;
		for other_tag in tags_iter {
			let other_val = match other_tag
				.text_by_field(field_name)?
				.map(|s| s.to_owned())
			{
				Some(val) => val,
				None => { // other tag doesn't have such field
					is_uniform = false;
					break;
				},
			};

			if first_val != other_val {
				is_uniform = false;
				break;
			}
		}

		Ok(if is_uniform { Some(first_val) } else { None })
	}
}
