use std::cell::RefCell;
use std::rc::Rc;
use winsafe::{self as w, prelude::*, gui};

use crate::ids;
use super::WndMain;

impl WndMain {
	#[must_use]
	pub fn new() -> w::AnyResult<Self> {
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_APP), Some(ids::ACC_MAIN));
		let lst_files = gui::ListView::new_dlg(&wnd, ids::LST_FILES, (H::Resize, V::Resize), Some(ids::MNU_MAIN));
		let tags = Rc::new(RefCell::new(Vec::default()));

		let new_self = Self { wnd, lst_files, tags };
		new_self.events();
		new_self.context_menu();
		Ok(new_self)
	}

	pub fn run(&self) -> w::AnyResult<i32> {
		self.wnd.run_main(None)
	}
}
