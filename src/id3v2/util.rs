use winsafe::{self as w, ErrResult};

const BOM_LE: u16 = 0xfe_ff;
const BOM_BE: u16 = 0xff_fe;

pub fn synch_safe_encode(mut n: u32) -> u32 { // big-endian
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

pub fn synch_safe_decode(n: u32) -> u32 { // big-endian
	let mut out: u32 = 0;
	let mut mask: u32 = 0x7f00_0000;

	while mask != 0 {
		out >>= 1;
		out |= n & mask;
		mask >>= 8;
	}

	out
}

pub fn parse_any_strings(src: &[u8]) -> ErrResult<Vec<String>> {
	match src[0] {
		0x00 => parse_iso88591_strings(&src[1..]), // skip encoding byte
		0x01 => parse_unicode_strings(&src[1..]),
		_ => Err(format!("Unrecognized text encoding: {}.", src[0]).into()),
	}
}

pub fn parse_iso88591_strings(src: &[u8]) -> ErrResult<Vec<String>> {
	let mut texts: Vec<String> = Vec::with_capacity(1); // arbitrary
	let mut buf16: Vec<u16> = Vec::default();

	for str_block in src.split(|b| *b == 0x00).into_iter() { // trailing zeros will be discarded
		buf16.clear();
		buf16.reserve(str_block.len());
		for ch in str_block.iter() {
			buf16.push(*ch as _); // simple expansion from u8 to u16, for each char
		}

		let parsed_str = w::WString::from_wchars_slice(&buf16);
		if !parsed_str.is_empty() {
			texts.push(parsed_str.to_string_checked()?);
		}
	}

	Ok(texts)
}

pub fn parse_unicode_strings(mut src: &[u8]) -> ErrResult<Vec<String>> {
	if (src.len() & 1) != 0 {
		// Length is not even, something is not quite right.
		// We'll simply discard the last byte and hope for the best.
		src = &src[..src.len() - 1];
	}

	// https://users.rust-lang.org/t/how-best-to-convert-u8-to-u16/57551/4
	let mut src16 = unsafe {
		std::slice::from_raw_parts(
			src.as_ptr().cast::<u16>(),
			src.len() / 2
		)
	};

	let mut is_little_endian = true;
	if src16[0] == BOM_LE || src16[0] == BOM_BE { // BOM found
		if src16[0] == BOM_BE { // big-endian
			is_little_endian = false;
		}
		src16 = &src16[1..]; // skip BOM
	}

	let mut texts = Vec::with_capacity(1); // arbitrary
	let mut buf16: Vec<u16> = Vec::default();

	for str_block in src16.split(|b| *b == 0x0000).into_iter() { // trailing zeros will be discarded
		buf16.clear();
		buf16.reserve(str_block.len());
		for ch in str_block.iter() {
			buf16.push(if is_little_endian { *ch } else { ch.swap_bytes() });
		}

		let parsed_str = w::WString::from_wchars_slice(&buf16);
		if !parsed_str.is_empty() {
			texts.push(parsed_str.to_string_checked()?);
		}
	}

	Ok(texts)
}

pub struct SerializedStrs {
	pub encoding_byte: u8,
	pub data: Vec<u8>,
}

impl SerializedStrs {
	pub fn new<S: AsRef<str>>(the_strings: &[S]) -> Self {
		let mut is_unicode = false;
		let mut estimated_len_bytes = 0;

		for one_string_ref in the_strings.iter() {
			let one_string = one_string_ref.as_ref();
			estimated_len_bytes += one_string.len(); // doesn't always bring the real char number, though

			is_unicode = one_string.chars().enumerate()
				.find(|(_, ch)| *ch as u32 > 255)
				.is_some();
		}

		if is_unicode {
			estimated_len_bytes *= 2; // will store as u16
			estimated_len_bytes += 2 * the_strings.len() - 1; // one zero char between each string
			estimated_len_bytes += 2; // BOM bytes
		} else {
			estimated_len_bytes += the_strings.len() - 1; // one zero char between each string
		}

		let mut buf: Vec<u8> = Vec::with_capacity(estimated_len_bytes);

		if is_unicode {
			buf.extend_from_slice(&BOM_LE.to_le_bytes()); // encode all Unicode strings as little-endian
		}

		for one_string_ref in the_strings.iter() {
			let one_string = one_string_ref.as_ref();
			for (_, ch) in one_string.chars().enumerate() {
				if is_unicode {
					buf.extend_from_slice(&(ch as u16).to_le_bytes());
				} else {
					buf.push(ch as u8);
				}
			}

			if is_unicode { // all strings are null-terminated
				buf.extend_from_slice(&[0x00, 0x00]);
			} else {
				buf.push(0x00);
			}
		}

		Self {
			encoding_byte: if is_unicode { 0x01 } else { 0x00 },
			data: buf,
		}
	}

	pub fn collect(&self) -> Vec<u8> {
		let mut buf: Vec<u8> = Vec::with_capacity(1 + self.data.len());
		buf.push(self.encoding_byte);
		buf.extend_from_slice(&self.data);
		buf
	}
}
