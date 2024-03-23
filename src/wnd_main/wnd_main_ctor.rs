use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::{self as w, prelude::*, co, gui};

use crate::{id3v2, ids};
use super::WndMain;

impl WndMain {
	/// Creates a new `WndMain` object.
	#[must_use]
	pub fn new() -> Self {
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_APP), Some(ids::ACC_MAIN));
		let lst_files = gui::ListView::new_dlg(&wnd, ids::LST_FILES, (H::Resize, V::Resize), Some(ids::MNU_MAIN));

		let new_self = Self {
			wnd,
			lst_files: lst_files.clone(),
			lst_files_h: gui::Header::from_list_view(&lst_files),
			all_tags: Rc::new(RefCell::new(Vec::default())),
		};
		new_self.wm_events();
		new_self.list_events();
		new_self
	}

	/// Runs the `WndMain` as the main application window.
	pub fn run(&self) -> w::AnyResult<i32> {
		self.wnd.run_main(None)
	}

	/// Initializes the `WndMain` window.
	pub(super) fn init_dialog(&self) -> w::AnyResult<bool> {
		self.update_num_files(self.lst_files.items().count());

		self.lst_files.set_extended_style(true, co::LVS_EX::FULLROWSELECT);
		self.lst_files.columns().add(&[
			("File", 380),
			("Pad", 50),
			("Pic", 30),
			("Artist", 160),
			("Title", 180),
			("Album", 180),
			("Track", 40),
			("Year", 40),
			("Genre", 100),
		]);

		Ok(true)
	}
}
