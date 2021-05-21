use std::error::Error;

pub fn synch_safe_encode(mut n: u32) -> u32 {
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

pub fn synch_safe_decode(n: u32) -> u32 {
	let mut out: u32 = 0;
	let mut mask: u32 = 0x7f00_0000;

	while mask != 0 {
		out >>= 1;
		out |= n & mask;
		mask >>= 8;
	}

	out
}

pub fn is_all_zero(blob: &[u8]) -> bool {
	for b in blob.iter() {
		if *b != 0x00 {
			return false;
		}
	}
	true
}

pub fn parse_any_strings(src: &[u8]) -> Result<Vec<String>, Box<dyn Error>> {
	match src[0] {
		0x00 => parse_iso88591_strings(src),
		0x01 => parse_unicode_strings(src),
		_ => Err(format!("Unrecognized text encoding: {}.", src[0]).into()),
	}
}

pub fn parse_iso88591_strings(src: &[u8]) -> Result<Vec<String>, Box<dyn Error>> {
	let mut texts = Vec::new();

	for str_block in src.split(|b| *b == 0x00).into_iter() {
		let parsed_str = std::str::from_utf8(str_block)?.to_string();
		if !parsed_str.is_empty() {
			texts.push(parsed_str);
		}
	}

	Ok(texts)
}

pub fn parse_unicode_strings(mut src: &[u8]) -> Result<Vec<String>, Box<dyn Error>> {
	if (src.len() & 1) != 0 {
		// Length is not even, something is not quite right.
		// We'll simply discard the last byte and hope for the best.
		src = &src[..src.len() - 1];
	}

	let mut texts = Vec::new();

	for str_block in src.split(|b| *b == 0x00).into_iter() {

	}

	Ok(texts)
}
