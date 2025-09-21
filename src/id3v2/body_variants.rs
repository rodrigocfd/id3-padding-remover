use winsafe::{self as w};

use crate::id3v2::*;

#[derive(Clone, PartialEq, Eq)]
pub struct Text {
	pub text: String,
}

impl std::fmt::Display for Text {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		write!(f, "{}", self.text)
	}
}

impl BodyVariant for Text {
	fn parse(src: &[u8]) -> w::AnyResult<Self> {
		let text = (|| {
			let (enc, src) = Enc::from_byte(src)?;
			let (info, _) = enc.parse_str(src)?;
			w::AnyResult::Ok(info)
		})()
		.map_err(|e| format!("Parse Text: {}", e.to_string()))?;

		Ok(Self { text })
	}

	fn serialize_size(&self) -> usize {
		let enc = Enc::from_serializing(&[&self.text]);
		let sz_text = enc.serialize_sz(&self.text);
		1 + sz_text
	}

	fn serialize(&self, dest: &mut Vec<u8>) {
		let enc = Enc::from_serializing(&[&self.text]);
		dest.push(enc as _);
		enc.serialize(dest, &self.text);
	}

	fn as_editable_str(&self) -> w::AnyResult<String> {
		Ok(self.text.clone())
	}

	fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		self.text = text.to_owned();
		Ok(())
	}
}

////////////////////////////////////////////////////////////////////////////////

#[derive(Clone, PartialEq, Eq)]
pub struct UserText {
	pub descr: String,
	pub text: String,
}

impl std::fmt::Display for UserText {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		if self.descr.trim().is_empty() {
			write!(f, "{}", self.text)
		} else {
			write!(f, "{} {}", self.descr, self.text)
		}
	}
}

impl BodyVariant for UserText {
	fn parse(src: &[u8]) -> w::AnyResult<Self> {
		let (descr, text) = (|| {
			let (enc, src) = Enc::from_byte(src)?;
			let (descr, src) = enc.parse_str(src)?;
			let (text, _) = enc.parse_str(src)?;
			w::AnyResult::Ok((descr, text))
		})()
		.map_err(|e| format!("Parse UserText: {}", e.to_string()))?;

		Ok(Self { descr, text })
	}

	fn serialize_size(&self) -> usize {
		let enc = Enc::from_serializing(&[&self.descr, &self.text]);
		let sz_descr = enc.serialize_sz(&self.descr);
		let sz_text = enc.serialize_sz(&self.text);
		1 + sz_descr + sz_text
	}

	fn serialize(&self, dest: &mut Vec<u8>) {
		let enc = Enc::from_serializing(&[&self.descr, &self.text]);
		dest.push(enc as _);
		enc.serialize(dest, &self.descr);
		enc.serialize(dest, &self.text);
	}

	fn as_editable_str(&self) -> w::AnyResult<String> {
		Ok(self.text.clone())
	}

	fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		self.descr = "".to_owned();
		self.text = text.to_owned();
		Ok(())
	}
}

////////////////////////////////////////////////////////////////////////////////

#[derive(Clone, PartialEq, Eq)]
pub struct Binary {
	pub data: Vec<u8>,
}

impl std::fmt::Display for Binary {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		write!(f, "{}", util::fmt_bytes(self.data.len()))
	}
}

impl BodyVariant for Binary {
	fn parse(src: &[u8]) -> w::AnyResult<Self> {
		Ok(Self { data: src.to_vec() }) // simply copy the bytes
	}

	fn serialize_size(&self) -> usize {
		self.data.len()
	}

	fn serialize(&self, dest: &mut Vec<u8>) {
		dest.extend_from_slice(&self.data);
	}

	fn as_editable_str(&self) -> w::AnyResult<String> {
		Err(format!("Cannot represent Binary as string.").into())
	}

	fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		Err(format!("Cannot represent Binary as string \"{text}\".").into())
	}
}

////////////////////////////////////////////////////////////////////////////////

#[derive(Clone, PartialEq, Eq)]
pub struct Comment {
	pub lang3: String,
	pub descr: String,
	pub text: String,
}

impl std::fmt::Display for Comment {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		if self.descr.trim().is_empty() {
			write!(f, "{}", self.text)
		} else {
			write!(f, "{} {}", self.descr, self.text)
		}
	}
}

impl BodyVariant for Comment {
	fn parse(src: &[u8]) -> w::AnyResult<Self> {
		let (lang3, descr, text) = (|| {
			let (enc, src) = Enc::from_byte(src)?;
			let (lang3, src) = util::parse_ascii(src, 3);
			let (descr, src) = enc.parse_str(src)?;
			let (text, _) = enc.parse_str(src)?;
			w::AnyResult::Ok((lang3, descr, text))
		})()
		.map_err(|e| format!("Parse Comment: {}", e.to_string()))?;

		Ok(Self { lang3, descr, text })
	}

	fn serialize_size(&self) -> usize {
		let enc = Enc::from_serializing(&[&self.descr, &self.text]);
		let sz_descr = enc.serialize_sz(&self.descr);
		let sz_text = enc.serialize_sz(&self.text);
		1 + 3 + sz_descr + sz_text
	}

	fn serialize(&self, dest: &mut Vec<u8>) {
		let enc = Enc::from_serializing(&[&self.descr, &self.text]);
		dest.push(enc as _);
		dest.extend(self.lang3.chars().map(|ch| ch as u8));
		enc.serialize(dest, &self.descr);
		enc.serialize(dest, &self.text);
	}

	fn as_editable_str(&self) -> w::AnyResult<String> {
		Ok(self.text.clone())
	}

	fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		self.descr = "".to_owned();
		self.text = text.to_owned();
		Ok(())
	}
}

////////////////////////////////////////////////////////////////////////////////

#[derive(Clone, PartialEq, Eq)]
pub struct Picture {
	pub mime: String,
	pub pic_type: PicType,
	pub descr: String,
	pub data: Vec<u8>,
}

impl std::fmt::Display for Picture {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		write!(f, "{} {}, {}", self.pic_type, self.mime, util::fmt_bytes(self.data.len()))
	}
}

impl BodyVariant for Picture {
	fn parse(src: &[u8]) -> w::AnyResult<Self> {
		let (mime, pic_type, descr, data) = (|| {
			let (enc, src) = Enc::from_byte(src)?;
			let (mime, src) = Enc::Iso88591.parse_str(src)?;

			let pic_type = PicType::try_from(src[0])?;
			let src = &src[1..];

			let (descr, src) = enc.parse_str(src)?;
			let data = src.to_vec();
			w::AnyResult::Ok((mime, pic_type, descr, data))
		})()
		.map_err(|e| format!("Parse Picture: {}", e.to_string()))?;

		Ok(Self { mime, pic_type, descr, data })
	}

	fn serialize_size(&self) -> usize {
		let enc = Enc::from_serializing(&[&self.descr]);
		let sz_mime = Enc::Iso88591.serialize_sz(&self.mime);
		let sz_descr = enc.serialize_sz(&self.descr);
		1 + sz_mime + 1 + sz_descr + self.data.len()
	}

	fn serialize(&self, dest: &mut Vec<u8>) {
		let enc = Enc::from_serializing(&[&self.descr]);
		dest.push(enc as _);
		Enc::Iso88591.serialize(dest, &self.mime);
		dest.push(self.pic_type as _);
		enc.serialize(dest, &self.descr);
		dest.extend_from_slice(&self.data);
	}

	fn as_editable_str(&self) -> w::AnyResult<String> {
		Err(format!("Cannot represent Picture as string.").into())
	}

	fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		Err(format!("Cannot represent Picture as string \"{text}\".").into())
	}
}

////////////////////////////////////////////////////////////////////////////////

#[derive(Clone, PartialEq, Eq)]
pub struct Geob {
	pub mime: String,
	pub filename: String,
	pub descr: String,
	pub enc_obj: Vec<u8>,
}

impl std::fmt::Display for Geob {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		write!(f, "{} {}, {}", self.mime, self.descr, util::fmt_bytes(self.enc_obj.len()))
	}
}

impl BodyVariant for Geob {
	fn parse(src: &[u8]) -> w::AnyResult<Self> {
		let (mime, filename, descr, enc_obj) = (|| {
			let (enc, src) = Enc::from_byte(src)?;
			let (mime, src) = Enc::Iso88591.parse_str(src)?;
			let (filename, src) = enc.parse_str(src)?;
			let (descr, src) = enc.parse_str(src)?;
			let enc_obj = src.to_vec();
			w::AnyResult::Ok((mime, filename, descr, enc_obj))
		})()
		.map_err(|e| format!("Parse Geob: {}", e.to_string()))?;

		Ok(Self { mime, filename, descr, enc_obj })
	}

	fn serialize_size(&self) -> usize {
		let enc = Enc::from_serializing(&[&self.descr]);
		let sz_mime = Enc::Iso88591.serialize_sz(&self.mime);
		let sz_filename = enc.serialize_sz(&self.filename);
		let sz_descr = enc.serialize_sz(&self.descr);
		1 + sz_mime + sz_filename + sz_descr + self.enc_obj.len()
	}

	fn serialize(&self, dest: &mut Vec<u8>) {
		let enc = Enc::from_serializing(&[&self.descr]);
		dest.push(enc as _);
		Enc::Iso88591.serialize(dest, &self.mime);
		enc.serialize(dest, &self.filename);
		enc.serialize(dest, &self.descr);
		dest.extend_from_slice(&self.enc_obj);
	}

	fn as_editable_str(&self) -> w::AnyResult<String> {
		Err(format!("Cannot represent Geob as string.").into())
	}

	fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		Err(format!("Cannot represent Geob as string \"{text}\".").into())
	}
}
