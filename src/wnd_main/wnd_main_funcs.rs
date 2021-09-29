use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::{self as w, gui, BoxResult};

use crate::id3v2::{FrameData, Tag};
use crate::ids::main as id;
use crate::util;
use super::{PreDelete, WndMain};

impl WndMain {
	pub fn new() -> BoxResult<Self> {
		let context_menu = w::HINSTANCE::NULL
			.LoadMenu(w::IdStr::Id(id::MNU_MAIN))?
			.GetSubMenu(0).unwrap();

		let wnd = gui::WindowMain::new_dlg(id::DLG_MAIN, Some(id::ICO_FROG), Some(id::ACT_MAIN));
		let lst_files = gui::ListView::new_dlg(&wnd, id::LST_FILES, Some(context_menu));
		let lst_frames = gui::ListView::new_dlg(&wnd, id::LST_FRAMES, None);
		let resizer = gui::Resizer::new(&wnd, &[
			(gui::Resz::Resize, gui::Resz::Resize, &[&lst_files]),
			(gui::Resz::Repos, gui::Resz::Resize, &[&lst_frames]),
		]);
		let tags_cache = Rc::new(RefCell::new(HashMap::default()));
		let app_name = util::app_name()?;

		let new_self = Self { wnd, lst_files, lst_frames, resizer, tags_cache, app_name };
		new_self.events();
		new_self.menu_events();
		Ok(new_self)
	}

	pub fn run(&self) -> BoxResult<i32> {
		self.wnd.run_main(None)
	}

	pub(super) fn titlebar_count(&self, moment: PreDelete) -> BoxResult<()> {
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

	pub(super) fn add_files<S: AsRef<str>>(&self, files: &[S]) -> BoxResult<()> {
		let clock = util::Timer::start()?;
		self.lst_files.set_redraw(false);

		for file_ref in files.iter() {
			let file = file_ref.as_ref();

			if self.lst_files.items().find(file).is_none() { // item not added yet?
				let tag = match Tag::read(file) { // parse the tag from file
					Ok(tag) => tag,
					Err(e) => {
						util::prompt::err(self.wnd.hwnd(),
							"Tag reading failed", Some("Error"),
							&format!("File: {}\n\n{}", file, e))?;
						return Ok(());
					},
				};

				self.lst_files.items().add(&[
					file,
					&format!("{}", tag.original_padding()), // write padding
				], Some(0))?;
				self.tags_cache.borrow_mut().insert(file.to_owned(), tag); // cache tag
			}
		}

		self.lst_files.set_redraw(true);
		self.lst_files.columns().set_width_to_fill(0)?;
		self.titlebar_count(PreDelete::No)?;

		util::prompt::info(self.wnd.hwnd(),
			"Operation successful", Some("Success"),
			&format!("{} file(s) loaded in {:.2} ms.", files.len(), clock.now_ms()?))?;
		Ok(())
	}

	pub(super) fn show_selected_tag_frames(&self) -> BoxResult<()> {
		let lvitems = self.lst_frames.items();
		lvitems.delete_all()?;

		let sel_files = self.lst_files.columns().selected_texts(0);
		self.lst_frames.hwnd().EnableWindow(!sel_files.is_empty());

		if sel_files.is_empty() { // nothing to do
			return Ok(());

		} else if sel_files.len() > 1 { // multiple selected items, just display a placeholder
			lvitems.add(&[
				"",
				&format!("{} selected...", sel_files.len()),
			], None)?;

		} else { // 1 single item selected, display its frames
			let tags_cache = self.tags_cache.borrow();
			let tag = tags_cache.get(&sel_files[0]).unwrap();

			for frame in tag.frames().iter() {
				let idx = lvitems.add(&[frame.name4()], None)?;

				match frame.data() {
					FrameData::Text(s) => lvitems.set_text(idx, 1, s)?,
					FrameData::MultiText(ss) => {
						lvitems.set_text(idx, 1, &ss[0])?;
						for i in 1..ss.len() {
							lvitems.add(&[
								"",
								&ss[i],
							], None)?;
						}
					},
					FrameData::Comment(com) => lvitems.set_text(idx, 1,
						&format!("[{}] {}", com.lang, com.text),
					)?,
					FrameData::Binary(bin) => lvitems.set_text(idx, 1,
						&format!("{} ({:.2}%)",
							&util::format_bytes(bin.len()),
							(bin.len() as f32) * 100.0 / tag.original_size() as f32),
					)?,
				}
			}
		}

		self.lst_frames.columns().set_width_to_fill(1)?;
		Ok(())
	}
}
