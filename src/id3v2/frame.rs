use winsafe::{self as w};

use crate::id3v2::*;

/// A unit of data within a tag.
#[derive(Clone, PartialEq, Eq)]
pub struct Frame {
	name4: String,
	flags: (u8, u8),
	body: Body,
}

impl std::fmt::Display for Frame {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		write!(f, "{}: {}", self.name4, self.body)
	}
}

impl Frame {
	#[must_use]
	pub const fn name4(&self) -> &String {
		&self.name4
	}
	#[must_use]
	pub const fn body(&self) -> &Body {
		&self.body
	}

	/// Constructs a new frame with dummy, invalida data.
	#[must_use]
	pub fn new_empty() -> Self {
		Self {
			name4: "TAAA".to_owned(),
			flags: (0, 0),
			body: Body::new_from_editable_str("TAAA", "").unwrap(),
		}
	}

	/// Constructs a new frame from an editable string. Only comment and text
	/// variants will work.
	#[must_use]
	pub(super) fn new_from_editable_str(name4: &str, val: &str) -> w::AnyResult<Self> {
		Ok(Self {
			name4: name4.to_owned(),
			flags: (0, 0),
			body: Body::new_from_editable_str(name4, val)?,
		})
	}

	/// Also returns declared size, including 10-byte frame header.
	#[must_use]
	pub(super) fn parse(src: &[u8]) -> w::AnyResult<(Self, usize)> {
		// Parse the 10-byte frame header.
		let (name4, _) = util::parse_ascii(src, 4);

		let mut declared_size = u32::from_be_bytes(src[4..8].try_into()?) as usize + 10; // also count 10-byte frame header
		if declared_size > src.len() {
			declared_size = src.len(); // if serialized with error, be complacent
		}

		let flags = (src[8], src[9]);

		// Skip frame header, truncate to declared frame size.
		let src = &src[10..declared_size];

		// Parse the frame contents.
		let body = Body::parse(&name4, src)?;

		Ok((Self { name4, flags, body }, declared_size))
	}

	/// Size in bytes of the body when serialized.
	#[must_use]
	pub(super) fn serialize_size(&self) -> usize {
		10 + self.body.serialize_size()
	}

	/// Serializes the body into a Vec.
	pub(super) fn serialize(&self, dest: &mut Vec<u8>) {
		util::serialize_ascii(dest, &self.name4);

		let sz_body = self.body.serialize_size();
		dest.extend((sz_body as u32).to_be_bytes()); // won't count 10-byte header

		dest.extend([self.flags.0, self.flags.1]);
		self.body.serialize(dest);
	}

	#[must_use]
	pub fn as_editable_str(&self) -> w::AnyResult<String> {
		self.body.as_editable_str()
	}

	pub fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		self.body.set_editable_str(text)
	}

	#[must_use]
	pub fn is_replay_gain(&self) -> bool {
		if self.name4 == "TXXX"
			&& let Body::UserText(f) = &self.body
		{
			return f.descr.starts_with("replaygain_");
		}
		false
	}
}
