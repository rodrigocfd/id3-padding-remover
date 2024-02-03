use super::consts::PicType;
use super::str_engine;

pub enum FrameData {
	Text(Text),
	UserText(UserText),
	Binary(Binary),
	Comment(Comment),
	Picture(Picture),
}

impl FrameData {
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
