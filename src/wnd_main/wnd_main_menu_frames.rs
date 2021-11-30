use winsafe::prelude::*;

use crate::util;
use super::{ids, TagOp, WndMain};

impl WndMain {
	pub(super) fn _menu_frames_events(&self) {
		self.wnd.on().wm_command_accel_menu(ids::MNU_FRAMES_REM, {
			let self2 = self.clone();
			move || {
				let clock = util::Timer::start()?;
				let sel_mp3 = self2.lst_mp3s.items().iter_selected()
					.next().unwrap().text(0); // assume there's only 1 selected MP3

				let frames_count = {
					let mut count = 0;
					let mut tags_cache = self2.tags_cache.lock().unwrap();
					let tag = tags_cache.get_mut(&sel_mp3).unwrap(); // the tag of the selected MP3

					for sel_item in self2.lst_frames.items()
						.iter_selected()
						.collect::<Vec<_>>()
						.iter()
						.rev()
					{
						if sel_item.text(0).is_empty() { continue; } // discard auxiliar lines

						let frame_index = sel_item.lparam()?; // sequential frame index within tag
						tag.frames_mut().remove(frame_index as _);
						count += 1;
					}

					count
				};

				if let Err(e) = self2._modal_tag_op(TagOp::SaveAndLoad, &[&sel_mp3]) {
					util::prompt::err(self2.wnd.hwnd(),
						"Error", Some("Frame removal failed"), &e.to_string())?;
				} else {
					self2._display_sel_tags_frames()?;
					self2.wnd_fields.feed(
						self2.lst_mp3s.items()
							.iter_selected()
							.map(|sel_mp3| sel_mp3.text(0))
							.collect::<Vec<_>>(),
					)?;
					util::prompt::info(self2.wnd.hwnd(),
						"Operation successful", Some("Success"),
						&format!("{} frame(s) removed from file in {:.2} ms.",
							frames_count, clock.now_ms()?))?;
				}

				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FRAMES_MOVE_UP, {
			let self2 = self.clone();
			move || {
				let sel_mp3 = self2.lst_mp3s.items().iter_selected()
					.next().unwrap().text(0); // assume there's only 1 selected MP3
				let frame_index = self2.lst_frames.items().iter_selected()
					.next().unwrap().lparam()? as usize; // assume there's only 1 selected (and valid) frame

				{
					let mut tags_cache = self2.tags_cache.lock().unwrap();
					let tag = tags_cache.get_mut(&sel_mp3).unwrap();
					tag.frames_mut().swap(frame_index, frame_index - 1); // assumes frame is not the first
				}

				if let Err(e) = self2._modal_tag_op(TagOp::SaveAndLoad, &[&sel_mp3]) {
					util::prompt::err(self2.wnd.hwnd(),
						"Error", Some("Frame move failed"), &e.to_string())?;
				} else {
					self2._display_sel_tags_frames()?;
				}

				Ok(())
			}
		});
	}
}
