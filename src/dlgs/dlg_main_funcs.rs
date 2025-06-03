use winsafe::{self as w, gui, prelude::*};

use super::{DlgMain, LIST_COLS};
use crate::{id3v2, msgbox};

impl DlgMain {
	pub(super) fn update_num_files(&self, tot_files: u32) -> w::SysResult<()> {
		let num_selec = self.lst_files.items().selected_count();
		self.wnd
			.hwnd()
			.SetWindowText(&format!("ID3 Fit ({}/{})", num_selec, tot_files))?;
		Ok(())
	}

	pub(super) fn sort_list(&self) -> w::AnyResult<()> {
		let (col, is_asc) = self.cur_sort.get(); // read current sort state
		let cols = self.lst_files.header().unwrap().items();
		cols.iter()?.for_each(|col| {
			col.set_arrow(gui::HeaderArrow::None); // remove arrow from all cols
		});

		if [1, 5, 8].contains(&col) {
			self.sort_numeric_col(col, is_asc)?; // padding, track no. or year
		} else {
			self.lst_files.items().sort(|a, b| {
				if is_asc { a.text(col).cmp(&b.text(col)) } else { b.text(col).cmp(&a.text(col)) }
			})?;
		}

		let new_arrow = if is_asc { gui::HeaderArrow::Asc } else { gui::HeaderArrow::Desc };
		cols.get(col).set_arrow(new_arrow); // draw arrow in current col
		Ok(())
	}

	fn sort_numeric_col(&self, num_col: u32, is_asc: bool) -> w::SysResult<()> {
		self.lst_files.items().sort(|a, b| {
			let text1 = a.text(num_col);
			let text2 = b.text(num_col);

			if let Ok(num1) = text1.parse::<u32>() {
				if let Ok(num2) = text2.parse::<u32>() {
					if is_asc {
						return num1.cmp(&num2);
					} else {
						return num2.cmp(&num1);
					}
				}
			}

			// One of the texts is not numeric, simply compare strings.
			if is_asc { text1.cmp(&text2) } else { text2.cmp(&text1) }
		})
	}

	pub(super) fn add_files_to_list(&self, file_paths: &[impl AsRef<str>]) -> w::AnyResult<()> {
		self.lst_files.set_redraw(false);

		file_paths
			.iter()
			.map(|file_path| file_path.as_ref())
			.try_for_each(|file_path| -> w::AnyResult<()> {
				if w::path::is_directory(file_path) {
					w::path::dir_walk(file_path).try_for_each(|inner_path| -> w::AnyResult<()> {
						let inner_path = inner_path?;
						if w::path::has_extension(&inner_path, &["mp3"]) {
							self.add_one_file_to_list(&inner_path)?;
						}
						Ok(())
					})
				} else if w::path::has_extension(file_path, &["mp3"]) {
					self.add_one_file_to_list(file_path)
				} else {
					// Should never happen; protected by UI.
					Err(format!("Not an MP3 file: {}", file_path).into())
				}
			})?;

		self.sort_list()?;
		self.lst_files.set_redraw(true);
		self.lst_files.cols().get(0).set_width_to_fill()?;
		self.update_num_files(self.lst_files.items().count())?;
		Ok(())
	}

	fn add_one_file_to_list(&self, file_path: &str) -> w::AnyResult<()> {
		let tag = id3v2::Tag::read_from_file(file_path)?; // load the tag from the MP3 file

		let item = match self.lst_files.items().find(file_path) {
			Some(existing_item) => {
				// MP3 already present in the list?
				let rc_tag = existing_item.data()?;
				*rc_tag.try_borrow_mut()? = tag; // replace the tag currently stored in the item
				existing_item
			},
			None => {
				// MP3 not yet in the list?
				let new_item = self.lst_files.items().add(&[file_path], Some(0), tag)?; // save tag in the item
				new_item
			},
		};

		Self::render_tag(item)?;
		Ok(())
	}

	pub(super) fn render_tag(item: gui::ListViewItem<'_, id3v2::Tag>) -> w::AnyResult<()> {
		let rc_tag = item.data()?; // retrieve tag saved in the listview item
		let tag = rc_tag.try_borrow()?;
		item.set_text(1, &tag.padding().to_string())?;
		item.set_text(2, if tag.frame_by_name4("APIC").is_some() { "✓" } else { "" })?;
		item.set_text(3, if tag.has_replay_gain() { "✓" } else { "" })?;

		LIST_COLS
			.iter()
			.enumerate()
			.skip(4) // columns 0-3 don't render actual tag frames
			.map(|(idx, (_, _, name4))| {
				let text = match tag.frame_by_name4(*name4) {
					Some(frame) => frame.body().to_string(), // tag has the frame rendered by this column
					None => "".to_owned(),                   // tag doesn't have this frame, render an empty string
				};
				(idx, text)
			})
			.try_for_each(|(idx, cell_text)| -> w::SysResult<()> {
				item.set_text(idx as _, &cell_text)?;
				Ok(())
			})?;

		Ok(())
	}

	pub(super) fn rename(&self, has_track_no: bool) -> w::AnyResult<()> {
		match self
			.lst_files
			.items()
			.iter_selected()
			.try_for_each(|sel_item| -> w::AnyResult<()> {
				let new_name = {
					let rc_tag = sel_item.data()?; // retrieve data saved in the listview item
					let tag = rc_tag.try_borrow()?;

					let mut new_name = String::with_capacity(30);
					if has_track_no {
						match tag.frame_by_name4("TRCK") {
							Some(track) => {
								new_name.push_str(&format!("{:0>2} ", track.as_editable_string()?));
							},
							None => return Err("Missing track field.".into()),
						}
					}
					match tag.frame_by_name4("TPE1") {
						Some(artist) => {
							new_name.push_str(&artist.as_editable_string()?);
						},
						None => return Err("Missing artist field.".into()),
					}
					match tag.frame_by_name4("TIT2") {
						Some(title) => {
							new_name.push_str(" - ");
							new_name.push_str(&title.as_editable_string()?);
						},
						None => return Err("Missing title field.".into()),
					}
					new_name.push_str(".mp3");
					new_name
				};

				let cur_path = sel_item.text(0);
				let new_path = w::path::replace_file_name(&cur_path, &new_name);
				w::MoveFile(&cur_path, &new_path)?;
				sel_item.set_text(0, &new_path)?;
				Ok(())
			}) {
			Err(e) => {
				msgbox::err(self.wnd.hwnd(), "Missing field(s)", None, &e.to_string())?;
			},
			Ok(_) => {
				self.sort_list()?;
			},
		}

		Ok(())
	}

	pub(super) fn remove_rg_art(&self, del_art: bool) -> w::AnyResult<()> {
		let sel_count = self.lst_files.items().selected_count();
		let window_title = if del_art { "Remove ReplayGain and art" } else { "Remove ReplayGain" };
		let content = format!(
			"Remove ReplayGain {} frames of {} tag{}?",
			if del_art { "and art" } else { "" },
			sel_count,
			if sel_count == 1 { "" } else { "s" },
		);

		if msgbox::ask(self.wnd.hwnd(), window_title, None, &content, "&Remove")? {
			self.lst_files.items().iter_selected().try_for_each(
				|sel_item| -> w::AnyResult<()> {
					{
						let rc_tag = sel_item.data()?; // retrieve tag saved in the listview item
						let mut tag = rc_tag.try_borrow_mut()?;
						tag.frames_mut().retain(|frame| !frame.is_replay_gain());
						if del_art {
							tag.frames_mut().retain(|frame| frame.name4() != "APIC");
						}
						tag.save_to_file(&sel_item.text(0))?; // save to MP3 file
					}
					Self::render_tag(sel_item)?;
					Ok(())
				},
			)?;
		}
		Ok(())
	}
}
