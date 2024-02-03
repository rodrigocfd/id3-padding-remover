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

		let texts = str_engine::parse_any(src)?;
		match texts.len() {
			0 => return Err("Comment frame has no texts.".into()),
			1 => {
				text = texts[0].clone(); // in case of 1 text, be lenient and assume empty description
			},
			2 => {
				descr = texts[0].clone();
				text = texts[1].clone();
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

		unimplemented!()

	}

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
