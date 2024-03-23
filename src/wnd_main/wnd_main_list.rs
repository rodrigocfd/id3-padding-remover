use winsafe::{self as w, prelude::*, co};

use crate::ids;
use super::WndMain;

impl WndMain {
	pub(super) fn list_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_menu_popup(move |p| {
			if self2.lst_files.context_menu().unwrap() == &p.hmenu {
				[ids::MNU_MAIN_EDIT, ids::MNU_MAIN_REMOVE].into_iter()
					.try_for_each(|id|
						p.hmenu.EnableMenuItem(
							w::IdPos::Id(id),
							self2.lst_files.items().selected_count() > 0, // at least 1 file selected?
						).map(|_| ())
					)?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_item_changed(move |_| {
			self2.update_num_files(self2.lst_files.items().count());
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_key_down(move |p| {
			if p.wVKey == co::VK::DELETE { // on DEL key, remove selected files from the list
				self2.remove_selected_files()?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_delete_item(move |_| {
			// Notification is sent before the list is updated.
			self2.update_num_files(self2.lst_files.items().count() - 1);
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files_h.on().hdn_item_click(move |p| {
			println!("Clicked {}, ID {}", p.iItem, self2.lst_files_h.ctrl_id());
			Ok(())
		});
	}
}
