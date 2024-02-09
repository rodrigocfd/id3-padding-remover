use winsafe::{prelude::*};

use super::WndMain;

impl WndMain {
	pub(super) fn update_num_files(&self) {
		let num_files = self.lst_files.items().count();
		let num_selec = self.lst_files.items().selected_count();
		self.wnd.set_text(&format!("ID3 Fit ({}/{})", num_selec, num_files));
	}
}
