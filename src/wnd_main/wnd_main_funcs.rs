use winsafe::{self as w, prelude::*, co, gui};

use crate::{id3v2, wnd_edit::WndEdit};
use super::WndMain;

impl WndMain {
	pub(super) fn update_num_files(&self, tot_files: u32) {
		let num_selec = self.lst_files.items().selected_count();
		self.wnd.set_text(&format!("ID3 Fit ({}/{})", num_selec, tot_files));
	}

	pub(super) fn add_files_to_list(&self, files: &[impl AsRef<str>]) -> w::AnyResult<()> {
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
							let new_item = self.lst_files.items().add(&[mp3_path], Some(0), tag); // save tag in the item
							new_item
						},
					};
					self.print_tag_in_listview(&item)?;
					w::AnyResult::Ok(())
				}
			})?;

		self.sort_list(0, true); // force re-sort by path
		self.lst_files.set_redraw(true);
		self.update_num_files(self.lst_files.items().count());
		Ok(())
	}

	pub(super) fn print_tag_in_listview(&self,
		item: &gui::spec::ListViewItem<'_, id3v2::Tag>,
	) -> w::AnyResult<()>
	{
		let rc_tag = item.data().unwrap(); // retrieve tag saved in the listview item
		let tag = rc_tag.try_borrow()?;
		item.set_text(1, &tag.padding().to_string());
		item.set_text(2, if tag.has_frame("APIC") { "✓" } else { "" });

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

		let rc_sel_tags = self.lst_files.items()
			.iter_selected()
			.map(|sel_item| sel_item.data().unwrap())
			.collect::<Vec<_>>();

		let wnd_edit = WndEdit::new(&self.wnd, rc_sel_tags)?;
		if wnd_edit.show()? {
			self.lst_files.set_redraw(false);
			self.lst_files.items()
				.iter_selected()
				.try_for_each(|sel_item| {
					self.print_tag_in_listview(&sel_item)?; // update the list with the new values

					let rc_tag = sel_item.data().unwrap(); // retrieve tag saved in the listview item
					rc_tag.try_borrow()?.save_to_file(&sel_item.text(0))?; // save to MP3 file

					w::AnyResult::Ok(())
				})?;
			self.lst_files.set_redraw(true);
		}
		Ok(())
	}

	pub(super) fn strip_replaygain_art(&self, strip_art: bool) -> w::AnyResult<()> {
		let sel_count = self.lst_files.items().selected_count();
		if self.wnd.hwnd().TaskDialog(
			Some(if strip_art { "Strip ReplayGain and art" } else { "Strip ReplayGain" }),
			None,
			Some(&format!("Strip ReplayGain {} frames of {} tag{}?",
				if strip_art { "and art" } else { "" },
				sel_count,
				if sel_count == 1 { "" } else { "s" },
			)),
			co::TDCBF::OK | co::TDCBF::CANCEL,
			w::IconRes::Warn,
		)? == co::DLGID::OK {
			self.lst_files.items()
				.iter_selected()
				.try_for_each(|sel_item| {
					{
						let rc_tag = sel_item.data().unwrap(); // retrieve tag saved in the listview item
						let mut tag = rc_tag.try_borrow_mut()?;
						tag.frames_mut().retain(|frame| !frame.is_replay_gain());
						if strip_art {
							tag.frames_mut().retain(|frame| frame.name4() != "APIC");
						}
						tag.save_to_file(&sel_item.text(0))?; // save to MP3 file
					}
					self.print_tag_in_listview(&sel_item)?;
					w::AnyResult::Ok(())
				})?;
		}
		Ok(())
	}

	pub(super) fn sort_list(&self, new_col: u32, force_asc: bool) {
		let (cur_col, reversed) = self.cur_sort_col.get();
		self.lst_files_h.items().get(cur_col).set_arrow(gui::HeaderArrow::None);

		if force_asc || new_col != cur_col {
			self.lst_files.items().sort(|a, b| a.text(new_col).cmp(&b.text(new_col)));
			self.lst_files_h.items().get(new_col).set_arrow(gui::HeaderArrow::Asc);
			self.cur_sort_col.set((new_col, false));
		} else {
			if !reversed {
				self.lst_files.items().sort(|a, b| b.text(new_col).cmp(&a.text(new_col)));
				self.lst_files_h.items().get(new_col).set_arrow(gui::HeaderArrow::Desc);
			} else {
				self.lst_files.items().sort(|a, b| a.text(new_col).cmp(&b.text(new_col)));
				self.lst_files_h.items().get(new_col).set_arrow(gui::HeaderArrow::Asc);
			}
			self.cur_sort_col.set((new_col, !reversed));
		}
	}
}
