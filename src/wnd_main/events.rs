use winsafe::{self as w, prelude::*, co, gui};

use crate::ids;
use super::WndMain;

impl WndMain {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.lst_files.set_extended_style(true, co::LVS_EX::FULLROWSELECT);
			self2.lst_files.columns().add(&[
				("File", 240),
				("Artist", 120),
				("Title", 120),
			]);

			Ok(true)
		});
	}
}
