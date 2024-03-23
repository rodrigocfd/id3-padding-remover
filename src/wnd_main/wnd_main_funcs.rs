use winsafe::{self as w, prelude::*, gui};

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

					let path_and_tag = id3v2::PathAndTag::read_from_file(mp3_path)?;
					let new_item = self.lst_files.items().add(&[mp3_path], None);
					self.write_tag_to_listview(new_item, &path_and_tag.tag);
					tags_ref.push(path_and_tag); // keep the tag

					w::AnyResult::Ok(())
				})?;
		}

		self.lst_files.set_redraw(true);
		self.update_num_files(self.lst_files.items().count());
		Ok(())
	}

	pub(super) fn write_tag_to_listview(&self, item: gui::spec::ListViewItem, tag: &id3v2::Tag) {
		item.set_text(1, &tag.padding().to_string());
		if tag.has_frame("APIC") { item.set_text(2, "✓"); }

		[id3v2::Field::Artist, id3v2::Field::Title, id3v2::Field::Album, id3v2::Field::Track,
			id3v2::Field::Year, id3v2::Field::Genre]
			.iter()
			.map(|field| tag.known_field(*field).unwrap_or_default())
			.enumerate()
			.for_each(|(idx, field_val)| item.set_text((idx as u32) + 3, &field_val));
	}

	pub(super) fn remove_selected_files(&self) -> w::AnyResult<()> {
		let sel_paths = self.lst_files.items()
			.iter_selected()
			.map(|item| item.text(0)).collect::<Vec<_>>();
		self.all_tags.try_borrow_mut()?
			.retain(|path_and_tag| !sel_paths.contains(&path_and_tag.mp3_path)); // remove from memory
		self.lst_files.items().delete_selected();
		Ok(())
	}
}
