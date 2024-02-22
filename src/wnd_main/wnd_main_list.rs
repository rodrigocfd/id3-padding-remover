use winsafe::{self as w, prelude::*};

use super::WndMain;

impl WndMain {
	pub(super) fn list_events(&self) {
		let self2 = self.clone();
		self.lst_files.on().lvn_item_changed(move |_| {
			self2.update_num_files();
			Ok(())
		});
	}
}
