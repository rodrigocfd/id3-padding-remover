use winsafe::{self as w};

const BOM_BE: u16 = 0xfeff;
const BOM_LE: u16 = 0xfffe;

/// Encoding byte. Parses and serializes strings.
#[derive(Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Enc {
	Iso88591 = 0x00,
	Unicode = 0x01,
}

impl std::fmt::Display for Enc {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		write!(
			f,
			"{}",
			match self {
				Enc::Iso88591 => "ISO-8859-1",
				Enc::Unicode => "Unicode",
			}
		)
	}
}

impl Enc {
	/// Returns the encoding byte and the post-byte src.
	#[must_use]
	pub fn from_byte(src: &[u8]) -> w::AnyResult<(Self, &[u8])> {
		let enc = match &src[0] {
			0x00 => Self::Iso88591,
			0x01 => Self::Unicode,
			n => return Err(format!("Unknown encoding: {n}.").into()),
		};
		Ok((enc, &src[1..]))
	}

	/// If at least one of the strings is Unicode, returns `Unicode`, otherwise
	/// `Iso88591`.
	#[must_use]
	pub fn from_serializing(strs: &[impl AsRef<str>]) -> Self {
		let last_ch = char::from_u32(0xff).unwrap(); // last char of ISO-8859-1 set
		let has_unicode_ch = strs
			.iter()
			.any(|str| str.as_ref().chars().any(|ch| ch > last_ch));
		if has_unicode_ch { Enc::Unicode } else { Enc::Iso88591 }
	}

	/// Parses the null-terminated string according to the encoding. Returns the
	/// string, and the post-string src.
	#[must_use]
	pub fn parse_str<'a>(&self, src: &'a [u8]) -> w::AnyResult<(String, &'a [u8])> {
		let idx_zero = src.iter().position(|by| *by == 0x00).unwrap_or(src.len());
		match self {
			Enc::Iso88591 => {
				let s = Self::parse_iso88591(&src[..idx_zero]);
				let src_past = &src[std::cmp::min(src.len(), idx_zero + 1)..];
				Ok((s, src_past))
			},
			Enc::Unicode => {
				if !idx_zero.is_multiple_of(2) {
					Err(format!("Odd number of bytes in Unicode string: {idx_zero}.").into())
				} else {
					let wsrc = unsafe {
						std::slice::from_raw_parts(src.as_ptr() as *const u16, idx_zero / 2)
					};
					let s = Self::parse_unicode(wsrc);
					let src_past = &src[std::cmp::min(src.len(), idx_zero + 2)..];
					Ok((s, src_past))
				}
			},
		}
	}

	#[must_use]
	fn parse_iso88591(src: &[u8]) -> String {
		if src.is_empty() {
			return String::new();
		}

		let mut wbuf = w::WString::new_alloc_buf(src.len()); // no need for terminating null
		src.iter()
			.zip(wbuf.as_mut_slice().iter_mut())
			.for_each(|(by, ch)| *ch = *by as _);
		wbuf.to_string()
	}

	#[must_use]
	fn parse_unicode(src: &[u16]) -> String {
		let (src, is_le) = match src[0] {
			BOM_BE => (&src[1..], false), // big-endian, skip BOM byte
			BOM_LE => (&src[1..], true),  // little-endian, skip BOM byte
			_ => (src, true),             // little-endian by default
		};

		if src.is_empty() {
			return String::new();
		}

		let mut wbuf = w::WString::new_alloc_buf(src.len()); // no need for terminating null
		src.iter()
			.zip(wbuf.as_mut_slice().iter_mut())
			.for_each(|(wo, ch)| {
				*ch = if is_le { wo.swap_bytes() } else { *wo };
			});
		wbuf.to_string()
	}

	/// Returns the number of bytes required to serialize the string as
	/// null-terminated, with the given encoding.
	#[must_use]
	pub fn serialize_sz(&self, s: &str) -> usize {
		match self {
			Enc::Iso88591 => s.chars().count() + 1, // plus terminating null
			Enc::Unicode => (s.chars().count() + 1 + 1) * 2, // plus BOM and terminating null
		}
	}

	/// Serializes the string into the destination buffer, according to the encoding.
	pub fn serialize(&self, dest: &mut Vec<u8>, s: &str) {
		match self {
			Enc::Iso88591 => {
				dest.extend(s.chars().map(|ch| ch as u8));
				dest.push(0x00); // terminating null
			},
			Enc::Unicode => {
				dest.extend_from_slice(&BOM_LE.to_le_bytes()); // insert BOM word; we serialize as little-endian
				dest.extend(
					s.chars()
						.flat_map(|ch| (ch as u16).to_le_bytes().into_iter()),
				);
				dest.extend(std::iter::repeat_n(0x00, 2)); // terminating null word
			},
		}
	}
}
