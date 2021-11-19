use winsafe::ErrResult;

use super::util;

/// The COMM frame type.
#[derive(PartialEq, Eq)]
pub struct FrameComment {
	pub lang:  String,
	pub descr: String, // usually an empty string
	pub text:  String,
}

impl FrameComment {
	pub fn new(lang: Option<&str>, descr: Option<&str>, text: &str) -> Self {
		Self {
			lang:  lang.unwrap_or("eng").to_owned(),
			descr: descr.unwrap_or_default().to_owned(),
			text:  text.to_owned(),
		}
	}

	pub fn parse(src: &[u8]) -> ErrResult<Self> {
		let mut src = src;
		if src[0] != 0x00 && src[1] != 0x01 {
			return Err(
				format!("Unrecognized comment encoding: {:#04x}.", src[0]).into(),
			);
		}
		let is_unicode = src[0] == 0x01;
		src = &src[1..]; // skip encoding byte

		// Retrieve 3-char language string, always ISO-8859-1.
		let lang = util::parse_strs::iso88591(&src[0..4])?.remove(0); // keep 1st string
		src = &src[3..];

		// Retrieve comment description and text.
		let texts = if is_unicode {
			util::parse_strs::unicode(src)?
		} else {
			util::parse_strs::iso88591(src)?
		};

		if texts.len() == 2 {
			Ok(Self { lang, descr: texts[0].clone(), text: texts[1].clone() })
		} else if texts.len() == 1 {
			Ok(Self { lang, descr: "".to_owned(), text: texts[0].clone() })
		} else {
			Err(
				format!("Comment frame with multiple texts: {}.", texts.len())
					.into(),
			)
		}
	}

	pub fn serialize_data(&self) -> Vec<u8> {
		let buf_text = util::SerializedStrs::new(&[&self.text]);

		let mut buf = Vec::<u8>::with_capacity(1 + 3 + buf_text.data.len());
		buf.push(buf_text.encoding_byte);
		buf.extend(
			self.lang.chars()
				.enumerate()
				.map(|(_, ch)| ch as u8),
		);
		buf.push(0x00);
		buf.extend_from_slice(&buf_text.data);

		buf
	}

	pub fn set_text(&mut self, text: &str) {
		self.text = text.to_owned();
	}
}
