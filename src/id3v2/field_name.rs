/// Known simple text fields.
#[derive(Clone, Copy, PartialEq, Eq)]
pub enum FieldName {
	Artist,
	Title,
	Subtitle,
	Album,
	Track,
	Year,
	Genre,
	Composer,
	Lyricist,
	OrigArtist,
	OrigAlbum,
	OrigYear,
	Performer,
	Comment, // behold: not a simple text frame
}

impl FieldName {
	pub fn names(self) -> (&'static str, &'static str) {
		match self {
			Self::Artist     => ("TPE1", "Artist"),
			Self::Title      => ("TIT2", "Title"),
			Self::Subtitle   => ("TIT3", "Subtitle"),
			Self::Album      => ("TALB", "Album"),
			Self::Track      => ("TRCK", "Track"),
			Self::Year       => ("TYER", "Year"),
			Self::Genre      => ("TCON", "Genre"),
			Self::Composer   => ("TCOM", "Composer"),
			Self::Lyricist   => ("TEXT", "Lyricist"),
			Self::OrigArtist => ("TOPE", "Original artist"),
			Self::OrigAlbum  => ("TOAL", "Original album"),
			Self::OrigYear   => ("TORY", "Original year"),
			Self::Performer  => ("TPE3", "Performer"),
			Self::Comment    => ("COMM", "Comment"),
		}
	}
}
