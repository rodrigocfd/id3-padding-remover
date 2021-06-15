use std::cell::RefCell;
use std::collections::HashMap;
use std::error::Error;
use std::rc::Rc;
use winsafe::{self as w, co, gui};

use crate::id3v2::{format_bytes, FrameData, Tag};
use crate::ids;
use super::PreDelete;
use super::WndMain;

impl WndMain {
	pub fn new() -> Self {
		let context_menu = w::HINSTANCE::NULL
			.LoadMenu(ids::MNU_MAIN).unwrap()
			.GetSubMenu(0).unwrap();

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_FROG), Some(ids::ACT_MAIN));
		let lst_files = gui::ListView::new_dlg(&wnd, ids::LST_FILES, Some(context_menu));
		let lst_frames = gui::ListView::new_dlg(&wnd, ids::LST_FRAMES, None);
		let resizer = gui::Resizer::new(&wnd, &[
			(gui::Resz::Resize, gui::Resz::Resize, &[&lst_files]),
			(gui::Resz::Repos, gui::Resz::Resize, &[&lst_frames]),
		]);
		let tags_cache = Rc::new(RefCell::new(HashMap::default()));

		let new_self = Self { wnd, lst_files, lst_frames, resizer, tags_cache };
		new_self.events();
		new_self.menu_events();
		new_self
	}

	pub fn run(&self) -> w::WinResult<()> {
		self.wnd.run_main(None)
	}

	pub fn titlebar_count(&self, moment: PreDelete) -> Result<(), Box<dyn Error>> {
		let lvitems = self.lst_files.items();
		let count = lvitems.count() - match moment {
			PreDelete::Yes => 1, // because LVN_DELETEITEM is fired before deletion
			PreDelete::No => 0,
		};
		self.wnd.hwnd().SetWindowText(
			&format!("{} ({}/{})", ids::TITLE, lvitems.selected_count(), count),
		)?;
		Ok(())
	}

	pub(super) fn add_files<S: AsRef<str>>(&self, files: &[S]) -> Result<(), Box<dyn Error>> {
		for file_ref in files.iter() {
			let file = file_ref.as_ref();

			if self.lst_files.items().find(file).is_none() { // item not added yet?
				let tag = match Tag::read(file) { // parse the tag from file
					Ok(tag) => tag,
					Err(e) => {
						self.msg_err("Tag reading failed", &format!("File: {}\n\n{}", file, e));
						return Ok(());
					},
				};

				let idx = self.lst_files.items().add(file, Some(0))?;
				self.lst_files.items().set_text(idx, 1, &format!("{}", tag.original_padding()))?; // write padding
				self.tags_cache.borrow_mut().insert(file.to_owned(), tag); // cache tag
			}
		}

		self.lst_files.columns().set_width_to_fill(0).unwrap();
		self.titlebar_count(PreDelete::No)
	}

	pub(super) fn show_tag_frames(&self) -> Result<(), Box<dyn Error>> {
		let lvitems = self.lst_frames.items();
		lvitems.delete_all().unwrap();

		let sel_files = self.lst_files.columns().selected_texts(0);
		self.lst_frames.hwnd().EnableWindow(!sel_files.is_empty());

		if sel_files.is_empty() { // nothing to do
			return Ok(());

		} else if sel_files.len() > 1 { // multiple selected items, just display a placeholder
			lvitems.add("", None).unwrap();
			lvitems.set_text(0, 1,
				&format!("{} selected...", sel_files.len())).unwrap();

		} else { // 1 single item selected, display its frames
			let tags_cache = self.tags_cache.borrow();
			let tag = tags_cache.get(&sel_files[0]).unwrap();

			for frame in tag.frames().iter() {
				let idx = lvitems.add(frame.name4(), None)?;

				match frame.data() {
					FrameData::Text(s) => lvitems.set_text(idx, 1, s)?,
					FrameData::MultiText(ss) => {
						lvitems.set_text(idx, 1, &ss[0])?;
						for i in 1..ss.len() {
							let sub_idx = lvitems.add("", None).unwrap();
							lvitems.set_text(sub_idx, 1, &ss[i]).unwrap();
						}
					},
					FrameData::Comment(com) => lvitems.set_text(idx, 1,
						&format!("[{}] {}", com.lang, com.text),
					)?,
					FrameData::Binary(bin) => lvitems.set_text(idx, 1,
						&format!("{} ({:.2}%)",
							&format_bytes(bin.len()),
							(bin.len() as f32) * 100.0 / tag.original_size() as f32),
					)?,
				}
			}
		}

		Ok(())
	}

	pub(super) fn write_selected_tags(&self) -> Result<(), Box<dyn Error>> {
		{
			let mut tags_cache = self.tags_cache.borrow_mut();

			for idx in self.lst_files.items().selected().iter() {
				let file = self.lst_files.items().text_str(*idx, 0);
				let tag = tags_cache.get_mut(&file).unwrap();
				tag.write(&file)?; // save tag to file, no padding is written

				*tag = Tag::read(&file)?; // load tag back from file
				self.lst_files.items().set_text(*idx, 1,
					&format!("{}", tag.original_padding()))?; // update padding info
			}
		}
		self.show_tag_frames()
	}

	pub(super) fn msg_info(&self, caption: &str, text: &str) {
		self.wnd.hwnd().TaskDialog(None, Some(ids::TITLE), Some(caption),
			Some(text), co::TDCBF::OK, w::IdTdicon::Tdicon(co::TD_ICON::INFORMATION),
		).unwrap();
	}

	pub(super) fn msg_err(&self, caption: &str, text: &str) {
		self.wnd.hwnd().TaskDialog(None, Some(ids::TITLE), Some(caption),
			Some(text), co::TDCBF::OK, w::IdTdicon::Tdicon(co::TD_ICON::ERROR),
		).unwrap();
	}
}