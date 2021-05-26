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
		let hmap = hfile.CreateFileMapping(None, co::PAGE::READONLY, None, None)?;
		let hview = hmap.MapViewOfFile(co::FILE_MAP::READ, 0, None)?;

		let mapped_slice = hview.as_slice(hfile.GetFileSizeEx()?);
		let tag = Self::parse(mapped_slice)?;

		hview.UnmapViewOfFile()?;
		hmap.CloseHandle()?;
		hfile.CloseHandle()?;

		Ok(tag)
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

	pub fn write(&self, file: &str) -> Result<(), Box<dyn Error>> {
		// let (hfile, _) = w::HFILE::CreateFile(file, co::GENERIC::READ | co::GENERIC::WRITE,
		// 	co::FILE_SHARE::NONE, None, co::DISPOSITION::OPEN_ALWAYS,
		// 	co::FILE_ATTRIBUTE::NORMAL, None)?;
		// let hmap = hfile.CreateFileMapping(None, co::PAGE::READWRITE, None, None)?;
		// let hview = hmap.MapViewOfFile(co::FILE_MAP::READ | co::FILE_MAP::WRITE, 0, None)?;


		// hfile.WriteFile(&self.serialize(), None)?;

		// hview.UnmapViewOfFile()?;
		// hmap.CloseHandle()?;
		// hfile.CloseHandle()?;
		Ok(())
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
		let mut frames = Vec::default();
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

	fn serialize(&self) -> Vec<u8> {
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
}
