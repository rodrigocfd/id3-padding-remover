use winsafe::{self as w, prelude::*, co, gui};

use super::WndMain;

impl WndMain {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.update_num_files();

			self2.lst_files.set_extended_style(true, co::LVS_EX::FULLROWSELECT);
			self2.lst_files.columns().add(&[
				("File", 380),
				("Artist", 160),
				("Title", 180),
				("Album", 180),
				("Track", 40),
				("Year", 40),
				("Genre", 100),
			]);

			Ok(true)
		});
	}
}
