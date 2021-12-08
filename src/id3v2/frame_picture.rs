use winsafe as w;

use super::{PicKind, util};

/// The APIC frame type.
#[derive(PartialEq, Eq)]
pub struct FramePicture {
	pub mime:      String,
	pub kind:      PicKind,
	pub descr:     String, // usually an empty string
	pub pic_bytes: Vec<u8>,
}

impl FramePicture {
	pub fn new(mime: &str, kind: PicKind, descr: Option<&str>, pic_bytes: &[u8]) -> Self {
		Self {
			mime: mime.to_owned(),
			kind,
			descr: descr.unwrap_or("").to_owned(),
			pic_bytes: pic_bytes.to_vec(),
		}
	}

	pub fn parse(src: &[u8]) -> w::ErrResult<Self> {
		let mut src = src;

		// MIME type.
		let encoding = src[0];
		let idx_first_zero = src[1..].iter().position(|ch| *ch == 0x00).unwrap() + 1;
		let mime = util::parse_strs::any(&src[..idx_first_zero])?[0].clone();
		src = &src[idx_first_zero + 1..];

		// Picture type.
		let kind = PicKind::from(src[0]);
		src = &src[1..];

		// Description.
		let idx_second_zero = src.iter().position(|ch| *ch == 0x00).unwrap();
		let descr = if encoding == 0x00 {
			util::parse_strs::iso88591(&src[..idx_second_zero])
		} else {
			util::parse_strs::unicode(&src[..idx_second_zero])
		}?[0].clone();
		src = &src[idx_second_zero + 1..];

		// Picture data itself.
		let pic_bytes = src.to_vec();

		Ok(Self { mime, kind, descr, pic_bytes })
	}

	pub fn serialize_data(&self) -> Vec<u8> {
		let serialized_mime = util::SerializedStrs::new(&[&self.mime]);
		let serialized_descr = util::SerializedStrs::new(&[&self.descr]);

		let mut buf = Vec::<u8>::with_capacity(
			1 + serialized_mime.str_z.len() + 1 + // encoding + picture kind
			serialized_descr.str_z.len() +
			self.pic_bytes.len(),
		);
		buf.extend_from_slice(&serialized_mime.collect());
		buf.push(self.kind as u8);
		buf.extend_from_slice(&serialized_descr.str_z); // no encoding byte
		buf.extend_from_slice(&self.pic_bytes);
		buf
	}
}
