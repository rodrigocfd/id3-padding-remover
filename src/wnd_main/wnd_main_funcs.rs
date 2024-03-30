use winsafe::{self as w, prelude::*, gui};

use crate::{id3v2, wnd_edit::WndEdit};
use super::WndMain;

impl WndMain {
	pub(super) fn update_num_files(&self, tot_files: u32) {
		let num_selec = self.lst_files.items().selected_count();
		self.wnd.set_text(&format!("ID3 Fit ({}/{})", num_selec, tot_files));
	}

	pub(super) fn add_files(&self, files: &[impl AsRef<str>]) -> w::AnyResult<()> {
		self.lst_files.set_redraw(false);

		files.iter()
			.map(|mp3_path| mp3_path.as_ref())
			.try_for_each(|mp3_path| {
				if !w::path::has_extension(mp3_path, &[".mp3"]) { // should never happen; protected by UI
					Err(format!("Not an MP3 file: {}", mp3_path).into())
				} else {
					let tag = id3v2::Tag::read_from_file(mp3_path)?; // load the tag from the MP3 file
					let item = match self.lst_files.items().find(mp3_path) {
						Some(item) => { // MP3 already present in the list?
							let rc_tag = item.data().unwrap();
							*rc_tag.borrow_mut() = tag; // replace the tag currently saved in the item
							item
						},
						None => { // MP3 not in the list yet?
							let new_item = self.lst_files.items().add(&[mp3_path], None, tag); // save tag in the item
							new_item
						},
					};
					self.write_tag_to_listview(&item)?;
					w::AnyResult::Ok(())
				}
			})?;

		self.lst_files.set_redraw(true);
		self.update_num_files(self.lst_files.items().count());
		Ok(())
	}

	pub(super) fn write_tag_to_listview(&self,
		item: &gui::spec::ListViewItem<'_, id3v2::Tag>,
	) -> w::AnyResult<()>
	{
		let rc_tag = item.data().unwrap(); // retrieve tag saved in the listview item
		let tag = rc_tag.try_borrow()?;
		item.set_text(1, &tag.padding().to_string());
		if tag.has_frame("APIC") { item.set_text(2, "✓"); }

		[id3v2::Field::Artist, id3v2::Field::Title, id3v2::Field::Album, id3v2::Field::Track,
			id3v2::Field::Year, id3v2::Field::Genre, id3v2::Field::Comment]
			.iter()
			.map(|field| tag.known_field(*field).unwrap_or_default())
			.enumerate()
			.for_each(|(idx, field_val)| item.set_text((idx as u32) + 3, &field_val));
		Ok(())
	}

	pub(super) fn edit_selected(&self) -> w::AnyResult<()> {
		if self.lst_files.items().selected_count() == 0 {
			return Ok(()); // Enter key will hit here even if there are no selected items
		}

		let rc_tags = self.lst_files.items()
			.iter_selected()
			.map(|sel_item| sel_item.data().unwrap())
			.collect::<Vec<_>>();

		let wnd_edit = WndEdit::new(&self.wnd, rc_tags);
		wnd_edit.show()?;

		self.lst_files.set_redraw(false);
		self.lst_files.items()
			.iter_selected()
			.try_for_each(|sel_item| {
				self.write_tag_to_listview(&sel_item)?; // update the list with the new values

				// let rc_tag = sel_item.data().unwrap(); // retrieve tag saved in the listview item
				// rc_tag.try_borrow()?.save_to_file(&sel_item.text(0))?; // save to MP3 file

				w::AnyResult::Ok(())
			})?;
		self.lst_files.set_redraw(true);
		Ok(())
	}
}
