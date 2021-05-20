use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe as w;
use winsafe::co;
use winsafe::gui;

use crate::id3v2::Tag;
use crate::ids;
use super::WndMain;

impl WndMain {
	pub fn new() -> Self {
		let context_menu = w::HINSTANCE::NULL
			.LoadMenu(w::IdStr::Id(ids::MNU_MAIN)).unwrap()
			.GetSubMenu(0).unwrap();

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_FROG), None);
		let lst_files = gui::ListView::new_dlg(&wnd, ids::LST_FILES, Some(context_menu));
		let lst_frames = gui::ListView::new_dlg(&wnd, ids::LST_FRAMES, None);
		let resizer = gui::Resizer::new(&wnd, &[
			(gui::Resz::Resize, gui::Resz::Resize, &[&lst_files, &lst_files]),
			(gui::Resz::Repos, gui::Resz::Resize, &[&lst_frames]),
		]);
		let tags = Rc::new(RefCell::new(HashMap::default()));

		let selfc = Self { wnd, lst_files, lst_frames, resizer, tags };
		selfc.events();
		selfc.menu_events();
		selfc
	}

	pub fn run(&self) -> w::WinResult<()> {
		self.wnd.run_main(None)
	}

	pub(super) fn add_files(&self, files: &Vec<String>) {
		for file in files.iter() {
			if self.lst_files.items().find(file).is_none() { // item not added yet?
				let tag = match Tag::read(file) {
					Ok(tag) => tag,
					Err(e) => {
						self.wnd.hwnd().MessageBox(
							&format!("Tag reading failed:\n{}\n\n{}", file, e),
							"Error",
							co::MB::ICONERROR,
						).unwrap();
						return
					},
				};

				self.tags.borrow_mut().insert(file.to_owned(), tag);
				self.lst_files.items().add(file, None).unwrap();
			}
		}
	}
}
