use winsafe::{self as w};

const BOM_LE: u16 = 0xfe_ff;
const BOM_BE: u16 = 0xff_fe;

pub mod synch_safe {
	pub fn encode(mut n: u32) -> u32 { // big-endian
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

	pub fn decode(n: u32) -> u32 { // big-endian
		let mut out: u32 = 0;
		let mut mask: u32 = 0x7f00_0000;

		while mask != 0 {
			out >>= 1;
			out |= n & mask;
			mask >>= 8;
		}

		out
	}
}

pub mod parse_strs {
	use super::{BOM_BE, BOM_LE, w};

	pub fn any(src: &[u8]) -> w::ErrResult<Vec<String>> {
		match src[0] {
			0x00 => iso88591(&src[1..]), // skip encoding byte
			0x01 => unicode(&src[1..]),
			_ => Err(format!("Unrecognized text encoding: {}.", src[0]).into()),
		}
	}

	pub fn iso88591(src: &[u8]) -> w::ErrResult<Vec<String>> {
		let mut src = src;
		if let Some(idx) = src.iter().rposition(|b| *b != 0x00) {
			src = &src[..=idx]; // right-trim zeros to avoid an extra empty string
		}

		let mut texts = Vec::<String>::with_capacity(2); // arbitrary
		let mut buf16 = Vec::<u16>::default();

		for str_block in src.split(|b| *b == 0x00).into_iter() { // trailing zeros will be discarded
			buf16.clear();
			buf16.reserve(str_block.len());
			str_block.iter()
				.for_each(|ch| buf16.push(*ch as _)); // simple expansion from u8 to u16, for each char

			texts.push( // note: empty strings will be added
				w::WString::from_wchars_slice(&buf16)
					.to_string_checked()?,
			);
		}

		Ok(texts)
	}

	pub fn unicode(src: &[u8]) -> w::ErrResult<Vec<String>> {
		let mut src = src;
		if (src.len() & 1) != 0 {
			// Length is not even, something is not quite right.
			// We'll simply discard the last byte and hope for the best.
			src = &src[..src.len() - 1];
		}

		// Cast &[u8] to &[u16].
		// https://users.rust-lang.org/t/how-best-to-convert-u8-to-u16/57551/4
		let mut src16 = unsafe {
			std::slice::from_raw_parts(
				src.as_ptr().cast::<u16>(),
				src.len() / 2,
			)
		};

		if let Some(idx) = src16.iter().rposition(|b| *b != 0x0000) {
			src16 = &src16[..=idx]; // right-trim zeros to avoid an extra empty string
		}

		let mut is_little_endian = true;
		if src16[0] == BOM_LE || src16[0] == BOM_BE { // BOM found
			if src16[0] == BOM_BE { // big-endian
				is_little_endian = false;
			}
			src16 = &src16[1..]; // skip BOM
		}

		let mut texts = Vec::<String>::with_capacity(2); // arbitrary
		let mut buf16 = Vec::<u16>::default();

		for str_block in src16.split(|b| *b == 0x0000).into_iter() { // trailing zeros will be discarded
			buf16.clear();
			buf16.reserve(str_block.len());
			str_block.iter()
				.for_each(|ch|
					buf16.push(if is_little_endian { *ch } else { ch.swap_bytes() }),
				);

			texts.push( // note: empty strings will be added
				w::WString::from_wchars_slice(&buf16)
					.to_string_checked()?,
			);
		}

		Ok(texts)
	}
}

/// Serializes strings into u8 vecs, null-terminated, along with their encoding.
pub struct SerializedStrs {
	pub encoding_byte: u8,
	pub str_z: Vec<u8>,
}

impl SerializedStrs {
	/// Creates a new object by serializing the given strings, null-terminated.
	pub fn new(the_strings: &[impl AsRef<str>]) -> Self {
		let mut is_unicode = false;
		let mut estimated_len_bytes = 0;

		for one_string in the_strings.iter().map(|s| s.as_ref()) {
			estimated_len_bytes += one_string.chars().count();

			let is_this_string_unicode = one_string.chars()
				.enumerate()
				.find(|(_, ch)| *ch as u32 > 127) // Mp3Tag appears to do this
				.is_some();
			if is_this_string_unicode { // at least one string is Unicode
				is_unicode = true;
				break;
			}
		}

		if is_unicode {
			estimated_len_bytes *= 2; // will store as u16
			estimated_len_bytes += 2 * the_strings.len(); // all strings are null-terminated
			estimated_len_bytes += 2; // BOM bytes
		} else {
			estimated_len_bytes += the_strings.len(); // all strings are null-terminated
		}

		let mut buf = Vec::<u8>::with_capacity(estimated_len_bytes);

		the_strings.iter()
			.map(|one_string| one_string.as_ref())
			.for_each(|one_string| {
				if is_unicode {
					// Append BOM bytes.
					// All strings will be encoded as little-endian.
					buf.extend_from_slice(&BOM_LE.to_le_bytes());
				}

				for (_, ch) in one_string.chars().enumerate() { // write each char of string
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
			});

		Self {
			encoding_byte: if is_unicode { 0x01 } else { 0x00 },
			str_z: buf,
		}
	}

	/// Returns the encoding byte and the serialized null-terminated strings in
	/// a single vec.
	pub fn collect(&self) -> Vec<u8> {
		let mut buf = Vec::<u8>::with_capacity(1 + self.str_z.len());
		buf.push(self.encoding_byte);
		buf.extend_from_slice(&self.str_z);
		buf
	}
}
