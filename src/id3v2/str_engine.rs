use winsafe::{self as w};

use super::consts::Enc;

const BOM_LE: u16 = 0xfeff;
const BOM_BE: u16 = 0xfffe;

/// Converts a simple non-null-terminated ASCII slice into a string.
#[must_use]
pub fn from_ascii(src: &[u8]) -> String {
	w::WString::from_wchars_slice(
		&src.iter()
			.map(|b| *b as u16)
			.chain(std::iter::once(0x0000))
			.collect::<Vec<_>>(),
	)
	.to_string()
}

/// Converts a string into simple non-null-terminated ASCII bytes.
#[must_use]
pub fn to_ascii(s: &str) -> Vec<u8> {
	s.chars().map(|ch| ch as u8).collect()
}

/// Parses one or more null-separated strings, ISO-8859-1 or Unicode.
#[must_use]
pub fn parse_any(src: &[u8]) -> w::AnyResult<Vec<String>> {
	match Enc::try_from(src[0])? {
		Enc::Iso88591 => parse_iso_88591(&src[1..]),
		Enc::Unicode => parse_unicode(&src[1..]),
	}
}

/// Parses one or more null-separated ISO-8859-1 strings.
#[must_use]
pub fn parse_iso_88591(src: &[u8]) -> w::AnyResult<Vec<String>> {
	let mut src = src;
	if let Some(idx) = src.iter().rposition(|b| *b != 0x00) {
		src = &src[..=idx]; // right-trim zeros to avoid an extra empty string
	}
	if src.is_empty() {
		return Ok(Vec::default()); // no strings
	}

	let mut buf16 = Vec::<u16>::default();
	let texts = src
		.split(|b| *b == 0x00)
		.map(|part| {
			if part.is_empty() {
				Ok(String::default()) // empty strings are also added
			} else {
				buf16.clear();
				buf16.extend(
					part.iter() // no need for a terminating null
						.map(|ch| *ch as u16), // simple expansion from u8 to u16, for each char
				);
				Ok(w::WString::from_wchars_slice(&buf16).to_string_checked()?)
			}
		})
		.collect::<w::AnyResult<Vec<_>>>()?;

	Ok(texts)
}

/// Parses one or more null-separated Unicode strings.
#[must_use]
pub fn parse_unicode(src: &[u8]) -> w::AnyResult<Vec<String>> {
	let mut src = src;
	if src.len() % 1 != 0 {
		// Length is not even, something is not quite right.
		// Discard last byte and hope for the best.
		src = &src[..src.len() - 1];
	}

	// Copy to buffer because slice::from_raw_parts() was crashing due to a
	// weird misalignment in some cases.
	let src16_buf = src
		.chunks(2)
		.map(|by| w::MAKEWORD(by[0], by[1]))
		.collect::<Vec<_>>();
	let mut src16 = src16_buf.as_slice();

	if let Some(idx) = src16.iter().rposition(|ch| *ch != 0x0000) {
		src16 = &src16[..=idx]; // right-trim zeros to avoid an extra empty string
	}
	if src16.is_empty() {
		return Ok(Vec::default()); // no strings
	}

	let mut buf16 = Vec::<u16>::default();
	let texts = src16
		.split(|ch| *ch == 0x0000)
		.map(|mut part| {
			let mut is_little_endian = true; // little-endian by default
			if part[0] == BOM_LE || part[0] == BOM_BE {
				if part[0] == BOM_BE {
					is_little_endian = false;
				}
				part = &part[1..]; // skip BOM
			}

			if part.is_empty() {
				Ok(String::default()) // empty strings are also added
			} else {
				buf16.clear();
				buf16.extend(
					part.iter() // no need for a terminating null
						.map(|ch| if is_little_endian { *ch } else { ch.swap_bytes() }),
				);
				Ok(w::WString::from_wchars_slice(&buf16).to_string_checked()?)
			}
		})
		.collect::<w::AnyResult<Vec<_>>>()?;

	Ok(texts)
}

/// Serializes the strings as null-terminated, returning the encoding byte and
/// the serialized bytes.
#[must_use]
pub fn serialize(strs: &[impl AsRef<str>]) -> (Enc, Vec<u8>) {
	let mut enc = Enc::Iso88591;
	let mut estimated_len_bytes = 0;

	for one_str in strs.iter().map(|s| s.as_ref()) {
		estimated_len_bytes += one_str.chars().count() + 1; // all strings will be null-terminated

		if enc == Enc::Iso88591 {
			// We still don't know if it's Unicode?
			let has_unicode_char = one_str.chars().position(|ch| ch as u32 > 0xff).is_some();
			if has_unicode_char {
				enc = Enc::Unicode; // at least 1 string is Unicode
			}
		}
	}

	if enc == Enc::Unicode {
		// Chars will be serialized as u16.
		estimated_len_bytes *= 2;
		estimated_len_bytes += 2 * strs.len(); // BOM bytes for each string
	}

	let mut ret_buf = Vec::<u8>::with_capacity(estimated_len_bytes);
	for one_str in strs.iter().map(|s| s.as_ref()) {
		if enc == Enc::Unicode {
			// Insert BOM bytes for each string.
			// Strings will be encoded as little-endian.
			ret_buf.extend(&BOM_LE.to_le_bytes());
		}

		for ch in one_str.chars() {
			// Write each char of the string.
			if enc == Enc::Unicode {
				ret_buf.extend(&(ch as u16).to_le_bytes()); // simple conversion to wide
			} else {
				ret_buf.push(ch as _); // simple narrowing to u8
			}
		}

		if enc == Enc::Unicode {
			ret_buf.extend(&[0x00, 0x00]); // append terminating null
		} else {
			ret_buf.push(0x00);
		}
	}

	(enc, ret_buf)
}
