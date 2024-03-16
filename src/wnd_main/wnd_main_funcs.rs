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

		{
			let mut tags_ref = self.all_tags.try_borrow_mut()?;
			files.iter()
				.map(|mp3_path| mp3_path.as_ref())
				.try_for_each(|mp3_path| {
					if !w::path::has_extension(mp3_path, &[".mp3"]) {
						return Err(format!("Not an MP3 file: {}", mp3_path).into());
					} else if self.lst_files.items().find(mp3_path).is_some() {
						return Ok(()); // ignore already existing files
					}

					let tag = id3v2::Tag::read_from_file(mp3_path)?;
					self.lst_files.items().add(&[
						mp3_path,
						&tag.known_field(id3v2::Field::Artist).unwrap_or_default(),
						&tag.known_field(id3v2::Field::Title).unwrap_or_default(),
						&tag.known_field(id3v2::Field::Album).unwrap_or_default(),
						&tag.known_field(id3v2::Field::Track).unwrap_or_default(),
						&tag.known_field(id3v2::Field::Year).unwrap_or_default(),
						&tag.known_field(id3v2::Field::Genre).unwrap_or_default(),
					], None);
					tags_ref.push(id3v2::PathAndTag::new(mp3_path, tag)); // keep the tag

					w::AnyResult::Ok(())
				})?;
		}

		self.lst_files.set_redraw(true);
		self.update_num_files(self.lst_files.items().count());
		Ok(())
	}
}
