use winsafe::{prelude::*, self as w, ErrResult};

use crate::id3v2::{FrameData, Tag, TextField};
use crate::util;
use super::{PreDelete, WndMain};

impl WndMain {
	pub(super) fn _titlebar_count(&self, moment: PreDelete) -> ErrResult<()> {
		let lv_items = self.lst_files.items();
		let count = lv_items.count() - match moment {
			PreDelete::Yes => 1, // because LVN_DELETEITEM is fired before deletion
			PreDelete::No => 0,
		};
		self.wnd.hwnd().SetWindowText(
			&format!("{} ({}/{})", self.app_name, lv_items.selected_count(), count),
		)?;
		Ok(())
	}

	pub(super) fn _add_files<S: AsRef<str>>(&self, files: &[S]) -> ErrResult<()> {
		self.lst_files.set_redraw(false);

		for file_ref in files.iter() {
			let file = file_ref.as_ref();

			let tag = match Tag::read(file) { // parse the tag from file
				Ok(tag) => tag,
				Err(e) => {
					util::prompt::err(self.wnd.hwnd(),
						"Tag reading failed", Some("Error"),
						&format!("File: {}\n\n{}", file, e))?;
					return Ok(()); // quit processing, nothing else is done
				}
			};

			if let Some(existing_item) = self.lst_files.items().find(file) { // file already in list?
				existing_item.set_text(1, &format!("{}", tag.original_padding()))?; // update padding
			} else {
				self.lst_files.items().add(&[ // insert new item
					file,
					&format!("{}", tag.original_padding()),
				], Some(0))?;
			}

			self.tags_cache.try_borrow_mut()?.insert(file.to_owned(), tag); // cache or re-cache the tag
		}

		self.lst_files.set_redraw(true);
		self.lst_files.columns().set_width_to_fill(0)?;
		self._titlebar_count(PreDelete::No)?;
		self._display_sel_tags_frames()?;
		Ok(())
	}

	pub(super) fn _display_sel_tags_frames(&self) -> ErrResult<()> {
		self.lst_frames.set_redraw(false);
		self.lst_frames.items().delete_all()?;

		let sel_count = self.lst_files.items().selected_count();

		if sel_count == 0 {
			// Nothing to do.

		} else if sel_count > 1 { // multiple selected items, just show a placeholder
			self.lst_frames.items().add(&[
				"",
				&format!("{} selected...", sel_count),
			], None)?;

		} else { // 1 single item selected, show its frames
			let sel_item = self.lst_files.items().iter_selected().next().unwrap();
			let tags_cache = self.tags_cache.try_borrow()?;
			let the_tag = tags_cache.get(&sel_item.text(0)).unwrap();

			for frame in the_tag.frames().iter() {
				let new_item = self.lst_frames.items().add(&[frame.name4()], None)?;

				match frame.data() {
					FrameData::Text(text) => new_item.set_text(1, text)?,
					FrameData::MultiText(texts) => {
						new_item.set_text(1, &texts[0])?;
						for text in texts.iter().skip(1) {
							self.lst_frames.items().add(&["", text], None)?; // add subsequent items
						}
					},
					FrameData::Comment(com) => {
						new_item.set_text(1, &format!("[{}] {}", com.lang, com.text))?;
					},
					FrameData::Binary(bin) => {
						new_item.set_text(1,
							&format!("{} ({:.2}%)",
								&util::format_bytes(bin.len()),
								(bin.len() as f32) * 100.0 / the_tag.original_size() as f32),
						)?
					},
				}
			}
		}

		self.lst_frames.set_redraw(true);
		self.lst_frames.hwnd().EnableWindow(sel_count > 0); // disable if no file selection
		self.lst_frames.columns().set_width_to_fill(1)?;
		Ok(())
	}

	pub(super) fn _remove_frames_from_sel_files_and_save(&self,
		replay_gain: bool, album_art: bool) -> ErrResult<()>
	{
		{
			let mut tags_cache = self.tags_cache.try_borrow_mut()?;

			for sel_item in self.lst_files.items().iter_selected() {
				let file_name = sel_item.text(0);
				let the_tag = tags_cache.get_mut(&file_name).unwrap();

				the_tag.frames_mut().retain(|frame| {
					if replay_gain && frame.name4() == "TXXX" {
						if let FrameData::MultiText(texts) = frame.data() {
							if texts[0].starts_with("replaygain_") {
								return false;
							}
						}
					}

					if album_art && frame.name4() == "APIC" {
						return false;
					}

					true
				});

				the_tag.write(&file_name)?; // save tag to file
			}
		}

		self._add_files( // reload all tags from their files
			&self.lst_files.items().iter_selected()
				.map(|item| item.text(0))
				.collect::<Vec<_>>(),
		)?;

		Ok(())
	}

	pub(super) fn _rename_files(&self, with_track: bool) -> ErrResult<()> {
		self.lst_files.set_redraw(false);
		let clock = util::Timer::start()?;
		let mut changed_count = 0;

		for sel_item in self.lst_files.items().iter_selected() {
			let file_name = sel_item.text(0);

			let new_name = {
				let tags_cache = self.tags_cache.try_borrow()?;
				let the_tag = tags_cache.get(&file_name).unwrap();

				let artist = util::clear_diacritics(
					the_tag.text_field(TextField::Artist).ok_or_else(|| "No artist frame.")??);
				let title = util::clear_diacritics(
					the_tag.text_field(TextField::Title).ok_or_else(|| "No title frame.")??);
				let track = util::clear_diacritics(
					the_tag.text_field(TextField::Track).ok_or_else(|| "No track frame.")??);

				w::Path::replace_file(&file_name,
					&if with_track {
						format!("{:02} {} - {}.mp3", track.parse::<u16>()?, artist, title)
					} else {
						format!("{} - {}.mp3", artist, title)
					})
			};
			if new_name == file_name { continue; } // same name, nothing to do

			{
				let mut tags_cache = self.tags_cache.try_borrow_mut()?;
				let renamed_tag = tags_cache.remove(&file_name).unwrap();
				tags_cache.insert(new_name.clone(), renamed_tag); // reinsert tag under new name
			}

			sel_item.set_text(0, &new_name)?; // change item
			w::MoveFile(&file_name, &new_name)?; // rename file on disk
			changed_count += 1;
		}

		self.lst_files.set_redraw(true);

		util::prompt::info(self.wnd.hwnd(),
			"Operation successful", Some("Success"),
			&format!("{} file(s) renamed in {:.2} ms.",
				changed_count, clock.now_ms()?))?;

		Ok(())
	}
}
