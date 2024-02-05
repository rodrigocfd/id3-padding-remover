use winsafe::{self as w, prelude::*, gui};

use crate::ids;
use super::WndMain;

impl WndMain {
	#[must_use]
	pub fn new() -> w::AnyResult<Self> {
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_APP), None);
		let lst_files = gui::ListView::new_dlg(&wnd, ids::LST_FILES, (H::Resize, V::Resize), None);

		let new_self = Self { wnd, lst_files };
		new_self.events();
		Ok(new_self)
	}

	pub fn run(&self) -> w::AnyResult<i32> {
		self.wnd.run_main(None)
	}
}
