use std::sync::{Arc, Mutex};
use winsafe::{prelude::*, self as w, path};

use crate::id3v2;
use crate::util;
use crate::wnd_progress::WndProgress;
use super::{PreDelete, WndMain};

impl WndMain {
	pub(super) fn _titlebar_count(&self, moment: PreDelete) -> w::ErrResult<()> {
		self.wnd.hwnd().SetWindowText(
			&format!("{} ({}/{})",
				self.app_name,
				self.lst_mp3s.items().selected_count(),
				self.lst_mp3s.items().count() - match moment {
					PreDelete::Yes => 1, // because LVN_DELETEITEM is fired before deletion
					PreDelete::No => 0,
				}),
		).map_err(|e| e.into())
	}

	pub(super) fn _add_files(&self, files: &[impl AsRef<str>]) -> w::ErrResult<()> {
		let process_err: Arc<Mutex<Option<w::ErrResult<()>>>>
			= Arc::new(Mutex::new(None)); // will receive any error from the processing closure

		WndProgress::new(&self.wnd, { // display the progress modal window
			let process_err = process_err.clone();
			let tags_cache = self.tags_cache.clone();
			let files = Arc::new(
				files.iter()
					.map(|s| s.as_ref().to_owned())
					.collect::<Vec<_>>(),
			);

			move || { // this closure will run in a spawned thread
				for file in files.iter() {
					let tag = match id3v2::Tag::read(file) { // read all files sequentially
						Ok(tag) => tag,
						Err(e) => {
							*process_err.lock().unwrap() = Some(Err(e)); // store error
							break; // no further files will be parsed
						},
					};
					tags_cache.lock().unwrap().insert(file.clone(), tag);
				}
				Ok(())
			}
		}).show()?;

		if let Some(e) = process_err.lock().unwrap().as_ref() {
			if let Err(e) = e {
				util::prompt::err(self.wnd.hwnd(),
					"Error", Some("Error parsing tag"), &e.to_string())?;
				return Ok(());
			}
		}

		self.lst_mp3s.set_redraw(false);

		for file in files.iter().map(|f| f.as_ref()) {
			let padding_txt = {
				let tags_cache = self.tags_cache.lock().unwrap();
				let tag = tags_cache.get(file).unwrap();
				if tag.is_empty() {
					"N/A".to_owned()
				} else {
					tag.padding().to_string()
				}
			};

			match self.lst_mp3s.items().find(file) {
				Some(item) => { item.set_text(1, &padding_txt)?; },
				None => { self.lst_mp3s.items().add(&[file], Some(0))?; }
			}
		}

		self.lst_mp3s.set_redraw(true);
		self.lst_mp3s.columns().set_width_to_fill(0)?;
		self._titlebar_count(PreDelete::No)?;
		Ok(())
	}

	pub(super) fn _display_sel_tags_frames(&self) -> w::ErrResult<()> {
		self.lst_frames.set_redraw(false);
		self.lst_frames.items().delete_all()?;

		let sel_count = self.lst_mp3s.items().selected_count();

		if sel_count == 0 {
			// Nothing to do.

		} else if sel_count > 1 { // multiple selected items, just show a placeholder
			self.lst_frames.items().add(&[
				"",
				&format!("{} selected...", sel_count),
			], None)?;

		} else { // 1 single item selected, show its frames
			let sel_item = self.lst_mp3s.items().iter_selected().next().unwrap();
			let tags_cache = self.tags_cache.lock().unwrap();
			let the_tag = tags_cache.get(&sel_item.text(0)).unwrap();

			for frame in the_tag.frames().iter() {
				use id3v2::FrameData;
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
								(bin.len() as f32) * 100.0 / the_tag.mp3_offset() as f32),
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
		replay_gain: bool, album_art: bool) -> w::ErrResult<()>
	{
		{
			let mut tags_cache = self.tags_cache.lock().unwrap();

			for sel_item in self.lst_mp3s.items().iter_selected() {
				let file_name = sel_item.text(0);
				let the_tag = tags_cache.get_mut(&file_name).unwrap();

				the_tag.frames_mut().retain(|frame| {
					if replay_gain && frame.name4() == "TXXX" {
						if let id3v2::FrameData::MultiText(texts) = frame.data() {
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
			&self.lst_mp3s.items().iter_selected()
				.map(|item| item.text(0))
				.collect::<Vec<_>>(),
		)?;

		Ok(())
	}

	pub(super) fn _rename_files(&self, with_track: bool) -> w::ErrResult<()> {
		// self.lst_files.set_redraw(false);
		// let clock = util::Timer::start()?;
		// let mut changed_count = 0;

		// for sel_item in self.lst_files.items().iter_selected() {
		// 	let file_name = sel_item.text(0);

		// 	let new_name = {
		// 		let tags_cache = self.tags_cache.try_borrow()?;
		// 		let the_tag = tags_cache.get(&file_name).unwrap();

		// 		let artist = util::clear_diacritics(
		// 			the_tag.text_field(FieldName::Artist).ok_or_else(|| "No artist frame.")??);
		// 		let title = util::clear_diacritics(
		// 			the_tag.text_field(FieldName::Title).ok_or_else(|| "No title frame.")??);
		// 		let track = util::clear_diacritics(
		// 			the_tag.text_field(FieldName::Track).ok_or_else(|| "No track frame.")??);

		// 		path::replace_file_name(&file_name,
		// 			&if with_track {
		// 				format!("{:02} {} - {}.mp3", track.parse::<u16>()?, artist, title)
		// 			} else {
		// 				format!("{} - {}.mp3", artist, title)
		// 			})
		// 	};
		// 	if new_name == file_name { continue; } // same name, nothing to do

		// 	{
		// 		let mut tags_cache = self.tags_cache.try_borrow_mut()?;
		// 		let renamed_tag = tags_cache.remove(&file_name).unwrap();
		// 		tags_cache.insert(new_name.clone(), renamed_tag); // reinsert tag under new name
		// 	}

		// 	sel_item.set_text(0, &new_name)?; // change item
		// 	w::MoveFile(&file_name, &new_name)?; // rename file on disk
		// 	changed_count += 1;
		// }

		// self.lst_files.set_redraw(true);

		// util::prompt::info(self.wnd.hwnd(),
		// 	"Operation successful", Some("Success"),
		// 	&format!("{} file(s) renamed in {:.2} ms.",
		// 		changed_count, clock.now_ms()?))?;

		Ok(())
	}
}
