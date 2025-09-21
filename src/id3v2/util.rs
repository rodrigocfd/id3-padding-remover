use winsafe::{self as w};

use crate::id3v2::*;

/// Encodes a big-endian number as synch-safe.
#[must_use]
pub const fn synchsafe_encode(mut n: u32) -> u32 {
	let mut out: u32 = 0;
	let mut mask: u32 = 0x7f;

	while (mask ^ 0x7fff_ffff) != 0 {
		out = n & !mask;
		out <<= 1;
		out |= n & mask;
		mask = ((mask + 1) << 8) - 1;
		n = out;
	}

	out
}

/// Decodes a big-endian number as synch-safe.
#[must_use]
pub const fn synchsafe_decode(n: u32) -> u32 {
	let mut out: u32 = 0;
	let mut mask: u32 = 0x7f00_0000;

	while mask != 0 {
		out >>= 1;
		out |= n & mask;
		mask >>= 8;
	}

	out
}

/// Converts a non-null-terminated ASCII slice into a string. Returns the
/// string, and the post-string src.
#[must_use]
pub fn parse_ascii(src: &[u8], num_chars: usize) -> (String, &[u8]) {
	let s = w::WString::from_wchars_slice(
		&src[..num_chars]
			.iter()
			.map(|b| *b as u16)
			.collect::<Vec<_>>(),
	)
	.to_string();

	(s, &src[num_chars..])
}

/// Writes the string as simple, non-null-terminated ASCII bytes.
pub fn serialize_ascii(dest: &mut Vec<u8>, s: &str) {
	dest.extend(s.chars().map(|ch| ch as u8))
}

/// More than 1,000 bytes will be converted to KB.
pub fn fmt_bytes(b: usize) -> String {
	if b > 1000 { format!("{:.1} KB", (b as f64) / 1000.0) } else { format!("{} bytes", b) }
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
