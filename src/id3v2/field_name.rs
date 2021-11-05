/// Known simple text fields.
#[derive(Clone, Copy, PartialEq, Eq)]
pub enum FieldName {
	Album,
	Artist,
	Genre,
	Title,
	Track,
	Year,
	Composer,
	Comment, // behold: not a simple text frame
}

impl FieldName {
	pub fn names(self) -> (&'static str, &'static str) {
		match self {
			Self::Album    => ("TALB", "Album"),
			Self::Artist   => ("TPE1", "Artist"),
			Self::Genre    => ("TCON", "Genre"),
			Self::Title    => ("TIT2", "Title"),
			Self::Track    => ("TRCK", "Track"),
			Self::Year     => ("TYER", "Year"),
			Self::Composer => ("TCOM", "Composer"),
			Self::Comment  => ("COMM", "Comment"),
		}
	}
}
