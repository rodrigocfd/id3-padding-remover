use std::error::Error;
use std::fmt;
use winsafe::co;

/// An error from an MP3 processing.
#[derive(Debug)]
pub struct Mp3Error {
	pub mp3_file: String,
	pub err: Box<dyn Error + Send + Sync>,
}

impl Error for Mp3Error {}

impl fmt::Display for Mp3Error {
	fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
		write!(f, "{}\n\n{}", self.mp3_file, self.err)
	}
}

impl From<co::ERROR> for Mp3Error {
	fn from(e: co::ERROR) -> Self {
		Self::new("", e.into())
	}
}

impl From<Box<dyn Error + Send + Sync>> for Mp3Error {
	fn from(e: Box<dyn Error + Send + Sync>) -> Self {
		Self::new("", e.into())
	}
}

impl Mp3Error {
	pub fn new(mp3_file: &str, err: Box<dyn Error + Send + Sync>) -> Self {
		Self {
			mp3_file: mp3_file.to_owned(),
			err,
		}
	}
}
