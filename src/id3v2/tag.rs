use std::error::Error;
use winsafe as w;
use winsafe::co;

use super::Frame;

pub struct Tag {
	frames: Vec<Frame>,
}

impl Tag {
	pub fn read(file: &str) -> Result<Self, Box<dyn Error>> {
		let (hfile, _) = w::HFILE::CreateFile(file, co::GENERIC::READ,
			co::FILE_SHARE::READ, None, co::DISPOSITION::OPEN_EXISTING,
			co::FILE_ATTRIBUTE::NORMAL, None)?;
		let bytes = hfile.ReadFile(hfile.GetFileSizeEx()? as _, None)?;
		hfile.CloseHandle()?;
		Self::parse(&bytes)
	}

	pub fn parse(src: &[u8]) -> Result<Self, Box<dyn Error>> {
		Err("NOT HERE YET".into())
	}

	pub fn frames(&self) -> &Vec<Frame> {
		&self.frames
	}
}
