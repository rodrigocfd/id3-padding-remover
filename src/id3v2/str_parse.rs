use winsafe::{self as w};

const BOM_LE: u16 = 0xfeff;
const BOM_BE: u16 = 0xfffe;

/// Parses one or more null-separated strings, ISO-8859-1 or Unicode.
pub fn any(src: &[u8]) -> w::AnyResult<Vec<String>> {
	match src[0] {
		0x00 => iso_88591(src),
		0x01 => unicode(src),
		_ => Err(format!("Unrecognized encoding: {}.", src[0]).into()),
	}
}

/// Parses one or more null-separated ISO-8859-1 strings.
pub fn iso_88591(src: &[u8]) -> w::AnyResult<Vec<String>> {
	let mut src = src;
	if let Some(idx) = src.iter().rposition(|b| *b != 0x00) {
		src = &src[..=idx]; // right-trim zeros to avoid an extra empty string
	}
	if src.is_empty() {
		return Ok(Vec::<String>::default());
	}

	let mut texts = Vec::<String>::with_capacity(2); // arbitrary
	let mut buf16 = Vec::<u16>::default();
	for part in src.split(|b| *b == 0x00) {
		if part.is_empty() {
			texts.push(String::default()); // empty strings are also added
		} else {
			buf16.clear();
			buf16.extend(
				part.iter()
					.map(|ch| *ch as u16) // simple expansion from u8 to u16, for each char
					.chain(std::iter::once(0x0000)), // terminating null
			);
			texts.push(w::WString::from_wchars_slice(&buf16).to_string_checked()?);
		}
	}
	Ok(texts)
}

/// Parses one or more null-separated Unicode strings.
pub fn unicode(src: &[u8]) -> w::AnyResult<Vec<String>> {
	let mut src = src;
	if src.len() % 1 != 0 {
		// Length is not even, something is not quite right.
		// Discard last byte and hope for the best.
		src = &src[..src.len() - 1];
	}

	let mut src16 = unsafe {
		std::slice::from_raw_parts(src.as_ptr() as *const u16, src.len() / 2)
	};
	if let Some(idx) = src16.iter().rposition(|ch| *ch != 0x0000) {
		src16 = &src16[..=idx]; // right-trim zeros to avoid an extra empty string
	}
	if src16.is_empty() {
		return Ok(Vec::<String>::default());
	}

	let mut texts = Vec::<String>::with_capacity(2); // arbitrary
	let mut buf16 = Vec::<u16>::default();
	for mut part in src16.split(|ch| *ch == 0x0000) {
		let mut is_little_endian = true; // little-endian by default
		if part[0] == BOM_LE || part[0] == BOM_BE {
			if part[0] == BOM_BE {
				is_little_endian = false;
			}
			part = &part[1..]; // skip BOM
		}

		if part.is_empty() {
			texts.push(String::default()); // empty strings are also added
		} else {
			buf16.clear();
			buf16.extend(
				part.iter()
					.map(|ch| if is_little_endian { *ch } else { ch.swap_bytes() })
					.chain(std::iter::once(0x0000)), // terminating null
			);
			texts.push(w::WString::from_wchars_slice(&buf16).to_string_checked()?);
		}
	}
	Ok(texts)
}
