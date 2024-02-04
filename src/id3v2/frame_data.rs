use winsafe::{self as w};

use super::consts::PicType;
use super::str_engine;

/// Polymorphic data of a frame.
pub enum FrameData {
	Text(Text),
	UserText(UserText),
	Binary(Binary),
	Comment(Comment),
	Picture(Picture),
}

impl FrameData {
	/// Parses the bytes according to the 4-char frame name.
	pub fn parse(name4: &str, src: &[u8]) -> w::AnyResult<Self> {
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

		let pic_type = PicType::from_u8(src[0]);
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
	pub fn serialize(&self) -> Vec<u8> {
		match self {
			FrameData::Text(t) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&t.text]);
				Vec::from_iter(
					std::iter::once(enc_byte)
						.chain(serialized.iter().map(|b| *b)),
				)
			},
			FrameData::UserText(ut) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&ut.descr, &ut.text]);
				Vec::from_iter(
					std::iter::once(enc_byte)
						.chain(serialized.iter().map(|b| *b)),
				)
			},
			FrameData::Binary(b) => b.data.clone(),
			FrameData::Comment(c) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&c.descr, &c.text]);
				Vec::from_iter(
					std::iter::once(enc_byte)
						.chain(c.lang3.chars().map(|ch| ch as u8))
						.chain(serialized.iter().map(|b| *b)),
				)
			},
			FrameData::Picture(p) => {
				let (enc_byte, serialized) = str_engine::serialize(&[&p.descr]);
				Vec::from_iter(
					std::iter::once(enc_byte)
						.chain(p.mime.chars().map(|ch| ch as u8))
						.chain(std::iter::once(0x00))
						.chain(std::iter::once(p.pic_type as u8))
						.chain(serialized.iter().map(|b| *b))
						.chain(p.data.iter().map(|b| *b)),
				)
			},
		}
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

pub struct Picture {
	pub mime: String,
	pub pic_type: PicType,
	pub descr: String,
	pub data: Vec<u8>,
}
