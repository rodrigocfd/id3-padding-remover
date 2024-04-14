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
				self2.lst_files.items().delete_selected();
			} else if p.wVKey == co::VK::RETURN { // on Enter key, edit the selected tags
				self2.edit_selected()?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().nm_dbl_clk(move |_| {
			self2.edit_selected()?;
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
			let new_col = p.iItem as u32;
			let cur_col = self2.cur_sort_col.get();
			if new_col == cur_col {
				self2.lst_files.items().sort(|a, b| b.text(new_col).cmp(&a.text(new_col))); // reverse order
				self2.cur_sort_col.set(0xffff_ffff); // won't matter
			} else {
				self2.lst_files.items().sort(|a, b| a.text(new_col).cmp(&b.text(new_col)));
				self2.cur_sort_col.set(new_col);
			}
			Ok(())
		});
	}
}
