use std::error::Error;

use super::util;

pub struct FrameComment {
	pub lang: String,
	pub text: String,
}

impl FrameComment {
	pub fn parse(mut src: &[u8]) -> Result<Self, Box<dyn Error>> {
		if src[0] != 0x00 && src[1] != 0x01 {
			return Err(
				format!("Unrecognized comment encoding: {:#04x}.", src[0]).into(),
			);
		}
		let is_unicode = src[0] == 0x01;
		src = &src[1..]; // skip encoding byte

		// Retrieve 3-char language string, always ISO-8859-1.
		let lang = std::str::from_utf8(&src[0..4])?.to_string();
		src = &src[3..];

		if src[0] == 0x00 {
			src = &src[1..]; // a null separator may appear, skip it
		}

		// Retrieve comment text.
		let texts = if is_unicode {
			util::parse_unicode_strings(src)?
		} else {
			util::parse_iso88591_strings(src)?
		};

		if texts.len() > 1 {
			return Err(
				format!("Comment frame with multiple texts: {}.",
					texts.len()).into(),
			);
		}

		Ok(Self { lang, text: texts[0].clone() })
	}
}
