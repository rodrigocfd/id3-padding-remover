use winsafe::{self as w, prelude::*, co};

use super::WndMain;

impl WndMain {
	pub(super) fn list_events(&self) {
		let self2 = self.clone();
		self.lst_files.on().lvn_item_changed(move |_| {
			self2.update_num_files(self2.lst_files.items().count());
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_key_down(move |p| {
			if p.wVKey == co::VK::DELETE { // on DEL key, remove selected files from the list
				self2.lst_files.items().delete_selected();
			}
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_delete_item(move |_| {
			// Notification is sent before the list is updated.
			self2.update_num_files(self2.lst_files.items().count() - 1);
			Ok(())
		});
	}
}
