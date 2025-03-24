use winsafe::{self as w};

use super::consts::PicType;
use super::str_engine;

/// Polymorphic data of a frame.
pub enum Body {
	Text(String),
	UserText(UserText),
	Binary(Vec<u8>),
	Comment(Comment),
	Picture(Picture),
}

pub struct UserText {
	pub descr: String,
	pub text: String,
}
pub struct Comment {
	pub lang3: String,
	pub descr: String,
	pub text: String,
}
pub struct Picture {
	pub mime: String,
	pub pic_type: PicType,
	pub descr: String,
	pub data: Vec<u8>,
}

impl std::fmt::Display for Body {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		use Body::*;
		write!(
			f,
			"{}",
			match self {
				Text(s) => s.clone(),
				UserText(ut) => format!("{} {}", ut.descr, ut.text),
				Binary(data) => format_bytes(data.len()),
				Comment(c) => c.text.clone(),
				Picture(p) => format!("{} {}, {}", p.pic_type, p.mime, format_bytes(p.data.len())),
			}
		)
	}
}

impl Eq for Body {}

impl PartialEq for Body {
	fn eq(&self, other: &Self) -> bool {
		use Body::*;
		match self {
			Text(s) => match other {
				Text(s2) => s == s2,
				_ => false,
			},
			UserText(ut) => match other {
				UserText(ut2) => ut.descr == ut2.descr && ut.text == ut2.text,
				_ => false,
			},
			Binary(data) => match other {
				Binary(data2) => data.iter().zip(data2.iter()).all(|(a, b)| a == b),
				_ => false,
			},
			Comment(c) => match other {
				Comment(c2) => c.lang3 == c2.lang3 && c.descr == c2.descr && c.text == c2.text,
				_ => false,
			},
			Picture(p) => match other {
				Picture(p2) => {
					p.mime == p2.mime
						&& p.pic_type == p2.pic_type
						&& p.descr == p2.descr
						&& p.data.iter().zip(p2.data.iter()).all(|(a, b)| a == b)
				},
				_ => false,
			},
		}
	}
}

impl Body {
	/// Creates a body from a string. If not possible, returns an error.
	#[must_use]
	pub(in crate::id3v2) fn new_from_string(name4: &str, val: &str) -> w::AnyResult<Self> {
		if name4 == "COMM" {
			Ok(Self::Comment(Comment {
				lang3: "eng".to_owned(), // default lang
				descr: "".to_owned(),    // assume blank description
				text: val.to_owned(),
			}))
		} else if name4.starts_with('T') {
			Ok(Self::Text(val.to_owned())) // simplest case
		} else {
			Err(format!("Cannot create a single-text {name4} frame.").into())
		}
	}

	/// Parses the bytes according to the 4-char frame name.
	#[must_use]
	pub(in crate::id3v2) fn parse(name4: &str, src: &[u8]) -> w::AnyResult<Self> {
		if name4 == "COMM" {
			Self::parse_comm(src)
		} else if name4 == "APIC" {
			Self::parse_apic(src)
		} else if name4.starts_with('T') {
			let texts = str_engine::parse_any(src)?;
			match texts.len() {
				0 => Err(format!("Frame {} contains no texts.", name4).into()),
				1 => Ok(Self::Text(texts[0].clone())),
				2 => Ok(Self::UserText(UserText {
					descr: texts[0].clone(),
					text: texts[1].clone(),
				})),
				_ => Err(format!("Frame {} contains {} texts.", name4, texts.len()).into()),
			}
		} else {
			// Anything else is treated as raw binary.
			Ok(Self::Binary(src.to_vec()))
		}
	}

	#[must_use]
	fn parse_comm(src: &[u8]) -> w::AnyResult<Self> {
		let mut src = src;
		let enc_byte = src[0];
		if enc_byte != 0x00 && enc_byte != 0x01 {
			return Err(format!("Unknown COMM encoding: {}.", enc_byte).into());
		}
		src = &src[1..]; // skip encoding byte

		let lang3 = str_engine::from_ascii(&src[..3]);
		src = &src[3..]; // skip lang chars

		let mut descr = String::default();
		let text;

		let mut texts = str_engine::parse_any(src)?;
		match texts.len() {
			0 => return Err("Comment frame has no texts.".into()),
			1 => {
				text = texts.remove(0); // in case of 1 text, be lenient and assume empty description
			},
			2 => {
				text = texts.remove(1);
				descr = texts.remove(0);
			},
			_ => return Err(format!("Comment frame has {} texts.", texts.len()).into()),
		}
		Ok(Self::Comment(Comment { lang3, descr, text }))
	}

	#[must_use]
	fn parse_apic(src: &[u8]) -> w::AnyResult<Self> {
		let mut src = src;
		let enc_byte = src[0];
		if enc_byte != 0x00 && enc_byte != 0x01 {
			return Err(format!("Unknown APIC encoding: {}.", enc_byte).into());
		}
		src = &src[1..]; // skip encoding byte

		let mut mime_parts = src.splitn(2, |b| *b == 0x00);
		let mime = str_engine::from_ascii(mime_parts.nth(0).unwrap()); // assume ASCII mime
		src = mime_parts.nth(0).unwrap();

		let pic_type = PicType::try_from(src[0])?;
		src = &src[1..]; // skip picture type

		let mut descr = String::default();
		if enc_byte == 0x00 {
			// Texts are ISO-8859-1.
			let mut descr_parts = src.splitn(2, |b| *b == 0x00);
			let mut texts = str_engine::parse_iso_88591(descr_parts.nth(0).unwrap())?;
			if texts.len() > 0 {
				descr = texts.remove(0); // description may be absent
			}
			src = descr_parts.nth(0).unwrap();
		} else {
			// Texts are Unicode.
			let idx_zero = src.windows(2).position(|bb| bb == &[0x00, 0x00]).unwrap();
			let mut texts = str_engine::parse_unicode(&src[..idx_zero])?;
			if texts.len() > 0 {
				descr = texts.remove(0); // description may be absent
			}
			src = &src[idx_zero + 2..];
		}

		Ok(Self::Picture(Picture {
			mime,
			pic_type,
			descr,
			data: src.to_vec(),
		}))
	}

	/// Serializes the data into bytes.
	#[must_use]
	pub(in crate::id3v2) fn serialize(&self) -> Vec<u8> {
		use Body::*;
		match self {
			Text(s) => {
				let (enc, serialized) = str_engine::serialize(&[&s]);
				std::iter::once(enc.into())
					.chain(serialized.iter().map(|b| *b))
					.collect()
			},
			UserText(ut) => {
				let (enc, serialized) = str_engine::serialize(&[&ut.descr, &ut.text]);
				std::iter::once(enc.into())
					.chain(serialized.iter().map(|b| *b))
					.collect()
			},
			Binary(data) => data.clone(),
			Comment(c) => {
				let (enc, serialized) = str_engine::serialize(&[&c.descr, &c.text]);
				std::iter::once(enc.into())
					.chain(c.lang3.chars().map(|ch| ch as u8))
					.chain(serialized.iter().map(|b| *b))
					.collect()
			},
			Picture(p) => {
				let (enc, serialized) = str_engine::serialize(&[&p.descr]);
				std::iter::once(enc.into())
					.chain(p.mime.chars().map(|ch| ch as u8))
					.chain(std::iter::once(0x00))
					.chain(std::iter::once(p.pic_type as u8))
					.chain(serialized.iter().map(|b| *b))
					.chain(p.data.iter().map(|b| *b))
					.collect()
			},
		}
	}

	/// Tries to set the value as a string, returning an error if not possible.
	pub(in crate::id3v2) fn set_string(&mut self, val: &str) -> w::AnyResult<()> {
		use Body::*;
		Ok(match self {
			Text(s) => {
				*s = val.to_owned();
			},
			UserText(ut) => {
				ut.text = val.to_owned();
			},
			Binary(_) => return Err("Binary data cannot be set as string.".into()),
			Comment(c) => {
				c.text = val.to_owned();
			},
			Picture(_) => return Err("Picture data cannot be set as string.".into()),
		})
	}
}

/// More than 1,000 bytes will be converted to KB.
fn format_bytes(b: usize) -> String {
	if b > 1000 { format!("{:.1} KB", (b as f64) / 1000.0) } else { format!("{} bytes", b) }
}
