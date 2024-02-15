use winsafe::{self as w, prelude::*};

use crate::id3v2;
use super::WndMain;

impl WndMain {
	pub(super) fn update_num_files(&self) {
		let num_files = self.lst_files.items().count();
		let num_selec = self.lst_files.items().selected_count();
		self.wnd.set_text(&format!("ID3 Fit ({}/{})", num_selec, num_files));
	}

	pub(super) fn add_files(&self, files: &[impl AsRef<str>]) -> w::AnyResult<()> {
		let self2 = self.clone();
		files.iter()
			.map(|file| file.as_ref())
			.try_for_each(|file| -> w::AnyResult<()> {
				if !w::path::has_extension(file, &[".mp3"]) {
					return Err(format!("Not an MP3 file: {}", file).into());
				} else if self2.lst_files.items().find(file).is_some() {
					return Ok(()); // ignore already existing files
				}

				let tag = id3v2::Tag::read_from_file(file)?;
				self2.lst_files.items().add(&[
					file,
					&tag.field(id3v2::Field::Artist).unwrap_or_default(),
					&tag.field(id3v2::Field::Title).unwrap_or_default(),
					&tag.field(id3v2::Field::Album).unwrap_or_default(),
					&tag.field(id3v2::Field::Track).unwrap_or_default(),
					&tag.field(id3v2::Field::Year).unwrap_or_default(),
					&tag.field(id3v2::Field::Genre).unwrap_or_default(),
				], None);

				Ok(())
			})?;

		self.update_num_files();
		Ok(())
	}
}
