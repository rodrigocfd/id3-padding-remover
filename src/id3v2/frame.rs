use winsafe::{self as w};

use super::body::Body;
use super::str_engine;

/// A unit of data within a tag.
#[derive(Clone, PartialEq, Eq)]
pub struct Frame {
	name4: String,
	flags: (u8, u8),
	body: Body,
}

impl Default for Frame {
	fn default() -> Self {
		Self {
			name4: "AAAA".to_owned(),
			flags: (0, 0),
			body: Body::Text("".to_owned()),
		}
	}
}

impl std::fmt::Display for Frame {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		write!(f, "{}: {}", self.name4, self.body)
	}
}

impl Frame {
	#[must_use]
	pub(in crate::id3v2) fn new_from_editable_string(name4: &str, val: &str) -> w::AnyResult<Self> {
		Ok(Self {
			name4: name4.to_owned(),
			flags: (0, 0),
			body: Body::new_from_editable_string(name4, val)?,
		})
	}

	/// Also returns declared size, including 10-byte frame header.
	#[must_use]
	pub(in crate::id3v2) fn parse(mut src: &[u8]) -> w::AnyResult<(Self, usize)> {
		// Parse the 10-byte frame header.
		let name4 = str_engine::from_ascii(&src[0..4]);
		let mut declared_size = u32::from_be_bytes(src[4..8].try_into()?) as usize + 10; // also count 10-byte frame header
		let flags = (src[8], src[9]);

		if declared_size > src.len() {
			declared_size = src.len(); // if serialized with error, be complacent
		}

		// Skip frame header, truncate to declared frame size.
		src = &src[10..declared_size];

		// Parse the frame contents.
		let body = Body::parse(&name4, src)?;

		Ok((Self { name4, flags, body }, declared_size))
	}

	#[must_use]
	pub(in crate::id3v2) fn serialize(&self) -> Vec<u8> {
		let serialized_body = self.body.serialize();
		str_engine::to_ascii(&self.name4)
			.into_iter()
			.chain((serialized_body.len() as u32).to_be_bytes()) // won't count 10-byte header
			.chain([self.flags.0, self.flags.1].into_iter())
			.chain(serialized_body.into_iter())
			.collect()
	}

	#[must_use]
	pub const fn name4(&self) -> &String {
		&self.name4
	}

	#[must_use]
	pub const fn body(&self) -> &Body {
		&self.body
	}

	#[must_use]
	pub fn as_editable_string(&self) -> w::AnyResult<String> {
		self.body.as_editable_string()
	}

	pub fn set_editable_string(&mut self, val: &str) -> w::AnyResult<()> {
		self.body.set_editable_string(val)
	}

	#[must_use]
	pub fn is_replay_gain(&self) -> bool {
		if self.name4 == "TXXX" {
			if let Body::UserText(f) = &self.body {
				return f.descr.starts_with("replaygain_");
			}
		}
		false
	}
}
