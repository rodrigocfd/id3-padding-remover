use std::sync::{Arc, Mutex};
use winsafe::{prelude::*, self as w};

use crate::id3v2;
use crate::util;
use crate::wnd_progress::WndProgress;
use super::{PreDelete, PrefixWithTrack, TagOp, WhatFrame, WndMain};

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

	pub(super) fn _modal_tag_op(&self,
		tag_op: TagOp,
		mp3_names: &[impl AsRef<str>]) -> w::ErrResult<()>
	{
		let process_err: Arc<Mutex<Option<w::ErrResult<()>>>>
			= Arc::new(Mutex::new(None)); // will receive any error from the processing closure

		WndProgress::new(&self.wnd, { // show the progress modal window
			let process_err = process_err.clone();
			let tags_cache = self.tags_cache.clone();
			let arc_mp3_names = Arc::new(
				mp3_names.iter()
					.map(|mp3_name| mp3_name.as_ref().to_owned())
					.collect::<Vec<_>>(),
			);

			move || { // this closure will run in a spawned thread
				for mp3_name in arc_mp3_names.iter() {
					if tag_op == TagOp::SaveAndLoad {
						let tags_cache = tags_cache.lock().unwrap();
						let cached_tag = tags_cache.get(mp3_name).unwrap();
						if let Err(e) = cached_tag.write(mp3_name) {
							*process_err.lock().unwrap() = Some(Err(e)); // store error
							break; // nothing else will be done
						}
					}

					let loaded_tag = match id3v2::Tag::read(mp3_name) {
						Ok(tag) => tag,
						Err(e) => {
							*process_err.lock().unwrap() = Some(Err(e)); // store error
							break; // nothing else will be done
						},
					};
					tags_cache.lock().unwrap().insert(mp3_name.clone(), loaded_tag); // store new loaded tag
				}
				Ok(())
			}
		}).show()?;

		if let Some(e) = process_err.lock().unwrap().take() {
			return e;
		}

		self.lst_mp3s.set_redraw(false);

		for mp3_name in mp3_names.iter().map(|m| m.as_ref()) { // add files + paddings on the listview
			let padding_txt = {
				let tags_cache = self.tags_cache.lock().unwrap();
				let tag = tags_cache.get(mp3_name).unwrap();
				if tag.is_empty() {
					"N/A".to_owned() // if the file has no tag, there's no padding
				} else {
					tag.padding().to_string()
				}
			};

			match self.lst_mp3s.items().find(mp3_name) {
				Some(item) => { item.set_text(1, &padding_txt)?; },
				None => { self.lst_mp3s.items().add(&[mp3_name, &padding_txt], Some(0))?; }
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

			for (idx, frame) in the_tag.frames().iter().enumerate() {
				use id3v2::FrameData;
				let new_item = self.lst_frames.items().add(&[frame.name4()], None)?;
				new_item.set_lparam(idx as _)?; // save frame index in item room

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

	pub(super) fn _remove_frames(&self,
		what: WhatFrame,
		mp3_names: &[impl AsRef<str>])
	{
		let mut tags_cache = self.tags_cache.lock().unwrap();
		mp3_names.iter()
			.map(|mp3_name| mp3_name.as_ref())
			.for_each(|mp3_name| {
				tags_cache.get_mut(mp3_name).unwrap()
					.frames_mut()
					.retain(|frame| {
						if frame.name4() == "TXXX" {
							if let id3v2::FrameData::MultiText(texts) = frame.data() {
								if texts[0].starts_with("replaygain_") {
									return false;
								}
							}
						}

						if what == WhatFrame::ReplgArt && frame.name4() == "APIC" {
							return false;
						}

						true
					});
			});
	}

	pub(super) fn _rename_files(&self, prefix: PrefixWithTrack) -> w::ErrResult<()> {
		self.lst_mp3s.set_redraw(false);
		let clock = util::Timer::start()?;
		let mut changed_count = 0;

		for sel_item in self.lst_mp3s.items().iter_selected() {
			let sel_mp3 = sel_item.text(0);

			let new_name = {
				let tags_cache = self.tags_cache.lock().unwrap();
				let the_tag = tags_cache.get(&sel_mp3).unwrap();

				let artist = util::clear_diacritics(
					the_tag.text_by_field(id3v2::FieldName::Artist)?.ok_or_else(|| "No artist frame.")?);
				let title = util::clear_diacritics(
					the_tag.text_by_field(id3v2::FieldName::Title)?.ok_or_else(|| "No title frame.")?);
				let track = util::clear_diacritics(
					the_tag.text_by_field(id3v2::FieldName::Track)?.ok_or_else(|| "No track frame.")?);

				w::path::replace_file_name(&sel_mp3,
					&if prefix == PrefixWithTrack::Yes {
						format!("{:02} {} - {}.mp3", track.parse::<u16>()?, artist, title)
					} else {
						format!("{} - {}.mp3", artist, title)
					})
			};
			if new_name == sel_mp3 { continue; } // same name, nothing to do

			{
				let mut tags_cache = self.tags_cache.lock().unwrap();
				let renamed_tag = tags_cache.remove(&sel_mp3).unwrap();
				tags_cache.insert(new_name.clone(), renamed_tag); // reinsert tag under new name
			}

			sel_item.set_text(0, &new_name)?; // change item
			w::MoveFile(&sel_mp3, &new_name)?; // rename file on disk
			changed_count += 1;
		}

		self.lst_mp3s.set_redraw(true);

		util::prompt::info(self.wnd.hwnd(),
			"Operation successful", Some("Success"),
			&format!("{} file(s) renamed in {:.2} ms.",
				changed_count, clock.now_ms()?))?;

		Ok(())
	}
}
