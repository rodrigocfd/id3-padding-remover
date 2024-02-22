use winsafe::{self as w, prelude::*};

use crate::id3v2;
use super::WndMain;

impl WndMain {
	pub(super) fn update_num_files(&self, tot_files: u32) {
		let num_selec = self.lst_files.items().selected_count();
		self.wnd.set_text(&format!("ID3 Fit ({}/{})", num_selec, tot_files));
	}

	pub(super) fn add_files(&self, files: &[impl AsRef<str>]) -> w::AnyResult<()> {
		self.lst_files.set_redraw(false);

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
					&tag.field(id3v2::Field::Artist),
					&tag.field(id3v2::Field::Title),
					&tag.field(id3v2::Field::Album),
					&tag.field(id3v2::Field::Track),
					&tag.field(id3v2::Field::Year),
					&tag.field(id3v2::Field::Genre),
				], None);

				Ok(())
			})?;

		self.lst_files.set_redraw(true);
		self.update_num_files(self.lst_files.items().count());
		Ok(())
	}
}
