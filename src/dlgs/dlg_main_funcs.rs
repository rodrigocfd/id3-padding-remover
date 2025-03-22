use winsafe::{self as w, co, gui, prelude::*};

use super::{DlgEdit, DlgMain, LIST_COLS};
use crate::id3v2;

impl DlgMain {
	pub(super) fn update_num_files(&self, tot_files: u32) -> w::SysResult<()> {
		let num_selec = self.lst_files.items().selected_count();
		self.wnd
			.hwnd()
			.SetWindowText(&format!("ID3 Fit ({}/{})", num_selec, tot_files))?;
		Ok(())
	}

	pub(super) fn add_files_to_list(&self, files: &[impl AsRef<str>]) -> w::AnyResult<()> {
		self.lst_files.set_redraw(false);

		files
			.iter()
			.map(|mp3_path| mp3_path.as_ref())
			.try_for_each(|mp3_path| {
				if !w::path::has_extension(mp3_path, &[".mp3"]) {
					// Should never happen; protected by UI.
					Err(format!("Not an MP3 file: {}", mp3_path).into())
				} else {
					let tag = id3v2::Tag::read_from_file(mp3_path)?; // load the tag from the MP3 file
					let item = match self.lst_files.items().find(mp3_path) {
						Some(item) => {
							// MP3 already present in the list?
							let rc_tag = item.data()?;
							*rc_tag.borrow_mut() = tag; // replace the tag currently saved in the item
							item
						},
						None => {
							// MP3 not yet in the list?
							let new_item = self.lst_files.items().add(&[mp3_path], Some(0), tag)?; // save tag in the item
							new_item
						},
					};
					Self::render_tag(item)?;
					w::AnyResult::Ok(())
				}
			})?;

		self.sort_list(0, true)?; // force re-sort by path
		self.lst_files.set_redraw(true);
		self.update_num_files(self.lst_files.items().count())?;
		Ok(())
	}

	pub(super) fn render_tag(item: gui::ListViewItem<'_, id3v2::Tag>) -> w::AnyResult<()> {
		let rc_tag = item.data()?; // retrieve tag saved in the listview item
		let tag = rc_tag.try_borrow()?;
		item.set_text(1, &tag.padding().to_string())?;
		item.set_text(2, if tag.frame("APIC").is_some() { "✓" } else { "" })?;
		item.set_text(3, if tag.has_replay_gain() { "✓" } else { "" })?;

		LIST_COLS
			.iter()
			.skip(4)
			.map(|(_, _, field)| match tag.frame(*field) {
				Some(frame) => frame.body().to_string(),
				None => "".to_owned(),
			})
			.enumerate()
			.try_for_each(|(idx, field_val)| {
				item.set_text((idx as u32) + 4, &field_val)?;
				w::SysResult::Ok(())
			})?;

		Ok(())
	}

	pub(super) fn edit_selected(&self) -> w::AnyResult<()> {
		if self.lst_files.items().selected_count() == 0 {
			return Ok(()); // Enter key will hit here even if there are no selected items
		}

		let rc_sel_tags = self
			.lst_files
			.items()
			.iter_selected()
			.map(|sel_item| sel_item.data())
			.collect::<w::SysResult<Vec<_>>>()?;

		let dlg_edit = DlgEdit::new(rc_sel_tags)?;
		if dlg_edit.show(&self.wnd)? {
			self.lst_files.set_redraw(false);
			self.lst_files
				.items()
				.iter_selected()
				.try_for_each(|sel_item| {
					Self::render_tag(sel_item)?; // update the list with the new values

					let rc_tag = sel_item.data()?; // retrieve tag saved in the listview item
					rc_tag.try_borrow()?.save_to_file(&sel_item.text(0))?; // save to MP3 file

					w::AnyResult::Ok(())
				})?;
			self.lst_files.set_redraw(true);
		}
		Ok(())
	}

	pub(super) fn strip_replaygain_art(&self, strip_art: bool) -> w::AnyResult<()> {
		let sel_count = self.lst_files.items().selected_count();
		let (res, _, _) = w::TaskDialogIndirect(&w::TASKDIALOGCONFIG {
			hwnd_parent: Some(self.wnd.hwnd()),
			window_title: Some(if strip_art {
				"Strip ReplayGain and art"
			} else {
				"Strip ReplayGain"
			}),
			main_icon: w::IconIdTd::Td(co::TD_ICON::WARNING),
			common_buttons: co::TDCBF::OK | co::TDCBF::CANCEL,
			flags: co::TDF::ALLOW_DIALOG_CANCELLATION | co::TDF::POSITION_RELATIVE_TO_WINDOW,
			content: Some(&format!(
				"Strip ReplayGain {} frames of {} tag{}?",
				if strip_art { "and art" } else { "" },
				sel_count,
				if sel_count == 1 { "" } else { "s" },
			)),
			..Default::default()
		})?;
		if res == co::DLGID::OK {
			self.lst_files
				.items()
				.iter_selected()
				.try_for_each(|sel_item| {
					{
						let rc_tag = sel_item.data()?; // retrieve tag saved in the listview item
						let mut tag = rc_tag.try_borrow_mut()?;
						tag.frames_mut().retain(|frame| !frame.is_replay_gain());
						if strip_art {
							tag.frames_mut().retain(|frame| frame.name4() != "APIC");
						}
						tag.save_to_file(&sel_item.text(0))?; // save to MP3 file
					}
					Self::render_tag(sel_item)?;
					w::AnyResult::Ok(())
				})?;
		}
		Ok(())
	}

	pub(super) fn sort_list(&self, new_col: u32, force_asc: bool) -> w::AnyResult<()> {
		let (cur_col, reversed) = self.cur_sort_col.get();
		let cols = self.lst_files.header().unwrap().items();

		cols.get(cur_col).set_arrow(gui::HeaderArrow::None); // remove arrow from current col

		if force_asc || new_col != cur_col {
			if [1, 5, 8].contains(&new_col) {
				self.sort_numeric_col(new_col, true)?; // padding, track no. or year
			} else {
				self.lst_files
					.items()
					.sort(|a, b| a.text(new_col).cmp(&b.text(new_col)))?;
			}
			cols.get(new_col).set_arrow(gui::HeaderArrow::Asc);
			self.cur_sort_col.set((new_col, false));
		} else {
			if !reversed {
				if [1, 5, 8].contains(&new_col) {
					self.sort_numeric_col(new_col, false)?; // padding, track no. or year
				} else {
					self.lst_files
						.items()
						.sort(|a, b| b.text(new_col).cmp(&a.text(new_col)))?;
				}
				cols.get(new_col).set_arrow(gui::HeaderArrow::Desc);
			} else {
				if [1, 5, 8].contains(&new_col) {
					self.sort_numeric_col(new_col, true)?; // padding, track no. or year
				} else {
					self.lst_files
						.items()
						.sort(|a, b| a.text(new_col).cmp(&b.text(new_col)))?; // reverse
				}
				cols.get(new_col).set_arrow(gui::HeaderArrow::Asc);
			}
			self.cur_sort_col.set((new_col, !reversed));
		}

		Ok(())
	}

	fn sort_numeric_col(&self, num_col: u32, asc: bool) -> w::SysResult<()> {
		self.lst_files.items().sort(|a, b| {
			let text1 = a.text(num_col);
			let text2 = b.text(num_col);

			if let Ok(num1) = text1.parse::<u32>() {
				if let Ok(num2) = text2.parse::<u32>() {
					if asc {
						return num1.cmp(&num2);
					} else {
						return num2.cmp(&num1);
					}
				}
			}

			// One of the texts is not numeric, simply compare strings.
			if asc { text1.cmp(&text2) } else { text2.cmp(&text1) }
		})
	}
}
