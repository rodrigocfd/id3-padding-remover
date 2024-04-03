use winsafe::{self as w};

use super::consts::{Field, PicType};
use super::str_engine;

/// Polymorphic data of a frame.
pub enum FrameData {
	Text(Text),
	UserText(UserText),
	Binary(Binary),
	Comment(Comment),
	Picture(Picture),
}

impl std::fmt::Display for FrameData {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
		use FrameData as F;
		write!(f, "{}", match self {
			F::Text(t) => t.text.clone(),
			F::UserText(ut) => ut.text.clone(),
			F::Binary(b) => format!("{} bytes", b.data.len()),
			F::Comment(c) => c.text.clone(),
			F::Picture(p) => format!("Pic {} {} bytes", p.mime, p.data.len()),
		})
	}
}

impl FrameData {
	/// Creates data from a string.
	#[must_use]
	pub(in crate::id3v2) fn new_from_string(f: Field, val: &str) -> Self {
		match f {
			Field::Comment => Self::Comment(Comment {
				lang3: "eng".to_owned(),
				descr: "".to_owned(), // will have empty description
				text: val.to_owned(),
			}),
			_ => Self::Text(Text {
				text: val.to_owned(),
			}),
		}
	}

	/// Parses the bytes according to the 4-char frame name.
	#[must_use]
	pub(in crate::id3v2) fn parse(name4: &str, src: &[u8]) -> w::AnyResult<Self> {
		if name4 == "COMM" {
			Self::parse_comm(src)
		} else if name4 == "APIC" {
			Self::parse_apic(src)
		} else if name4.starts_with("T") {
			let texts = str_engine::parse_any(src)?;
			match texts.len() {
				0 => Err(format!("Frame {} contains no texts.", name4).into()),
				1 => Ok(Self::Text(Text { text: texts[0].clone() })),
				2 => Ok(Self::UserText(UserText { descr: texts[0].clone(), text: texts[1].clone() })),
				_ => Err(format!("Frame {} contains {} texts.", name4, texts.len()).into()),
			}
		} else { // anything else is treated as raw binary
			Ok(Self::Binary(Binary { data: src.to_vec() }))
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

		let pic_type = PicType::from(src[0]);
		src = &src[1..]; // skip picture type

		let mut descr = String::default();
		if enc_byte == 0x00 { // ISO-8859-1
			let mut descr_parts = src.splitn(2, |b| *b == 0x00);
			let mut texts = str_engine::parse_iso_88591(descr_parts.nth(0).unwrap())?;
			if texts.len() > 0 { // description may be absent
				descr = texts.remove(0);
			}
			src = descr_parts.nth(0).unwrap();
		} else { // Unicode
			let idx_zero = src.windows(2).position(|bb| bb == &[0x00, 0x01]).unwrap();
			let mut texts = str_engine::parse_unicode(&src[..idx_zero])?;
			if texts.len() > 0 { // description may be absent
				descr = texts.remove(0);
			}
			src = &src[idx_zero + 1..];
		}

		Ok(Self::Picture(Picture { mime, pic_type, descr, data: src.to_vec() }))
	}

	/// Serializes the data into bytes.
	#[must_use]
	pub(in crate::id3v2) fn serialize(&self) -> Vec<u8> {
		match self {
			FrameData::Text(t) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&t.text]);
				std::iter::once(enc_byte)
					.chain(serialized.iter().map(|b| *b))
					.collect()
			},
			FrameData::UserText(ut) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&ut.descr, &ut.text]);
				std::iter::once(enc_byte)
					.chain(serialized.iter().map(|b| *b))
					.collect()
			},
			FrameData::Binary(b) => b.data.clone(),
			FrameData::Comment(c) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&c.descr, &c.text]);
				std::iter::once(enc_byte)
					.chain(c.lang3.chars().map(|ch| ch as u8))
					.chain(serialized.iter().map(|b| *b))
					.collect()
			},
			FrameData::Picture(p) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&p.descr]);
				std::iter::once(enc_byte)
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
		Ok(match self {
			FrameData::Text(t) => {
				t.text = val.to_owned();
			},
			FrameData::UserText(ut) => {
				ut.text = val.to_owned();
			},
			FrameData::Binary(_) => return Err("Binary data cannot be set as string.".into()),
			FrameData::Comment(c) => {
				c.text = val.to_owned();
			},
			FrameData::Picture(_) => return Err("Picture data cannot be set as string.".into()),
		})
	}
}

pub struct Text {
	pub text: String,
}

pub struct UserText {
	pub descr: String,
	pub text: String,
}

pub struct Binary {
	pub data: Vec<u8>,
}

pub struct Comment {
	pub lang3: String,
	pub descr: String,
	pub text: String,
}

#[derive(Eq)]
pub struct Picture {
	pub mime: String,
	pub pic_type: PicType,
	pub descr: String,
	pub data: Vec<u8>,
}

impl PartialEq for Picture {
	fn eq(&self, other: &Picture) -> bool {
		self.mime == other.mime
			&& self.pic_type == other.pic_type
			&& self.descr == other.descr
			&& self.data.iter()
				.zip(other.data.iter())
				.all(|(a, b)| a == b)
	}
}
