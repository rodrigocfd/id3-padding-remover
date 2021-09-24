use winsafe::BoxResult;

use super::tag_util;

/// The COMM frame type.
pub struct FrameComment {
	pub lang: String,
	pub text: String,
}

impl FrameComment {
	pub fn parse(mut src: &[u8]) -> BoxResult<Self> {
		if src[0] != 0x00 && src[1] != 0x01 {
			return Err(
				format!("Unrecognized comment encoding: {:#04x}.", src[0]).into(),
			);
		}
		let is_unicode = src[0] == 0x01;
		src = &src[1..]; // skip encoding byte

		// Retrieve 3-char language string, always ISO-8859-1.
		let lang = tag_util::parse_iso88591_strings(&src[0..4])?.remove(0);
		src = &src[3..];

		if src[0] == 0x00 {
			src = &src[1..]; // a null separator may appear, skip it
		}

		// Retrieve comment text.
		let texts = if is_unicode {
			tag_util::parse_unicode_strings(src)?
		} else {
			tag_util::parse_iso88591_strings(src)?
		};

		if texts.len() > 1 {
			return Err(
				format!("Comment frame with multiple texts: {}.",
					texts.len()).into(),
			);
		}

		Ok(Self { lang, text: texts[0].clone() })
	}

	pub fn serialize_data(&self) -> Vec<u8> {
		let buf_text = tag_util::SerializedStrs::new(&[&self.text]);

		let mut buf: Vec<u8> = Vec::with_capacity(1 + 3 + buf_text.data.len());
		buf.push(buf_text.encoding_byte);
		buf.extend(
			self.lang.chars().enumerate()
				.map(|(_, ch)| ch as u8),
		);
		buf.push(0x00);
		buf.extend_from_slice(&buf_text.data);

		buf
	}
}
