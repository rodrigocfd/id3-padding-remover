use winsafe::{prelude::*, ErrResult};

use super::WndFields;

impl WndFields {
	pub(super) fn _update_after_check(&self) -> ErrResult<()> {
		for field in self.fields.iter() {
			field.txt.hwnd().EnableWindow(field.chk.is_checked()); // enable field if its checkbox is on
		}

		// let sel_files_count = self.sel_files.try_borrow()?.len();
		// let at_least_1_check = self.fields
		// 	.iter().find(|field| field.chk.is_checked()).is_some();

		// self.btn_save.hwnd().EnableWindow(sel_files_count > 0 && at_least_1_check);
		// if at_least_1_check {
		// 	self.btn_save.hwnd().SetWindowText(&format!("&Save ({})", sel_files_count))
		// } else {
		// 	self.btn_save.hwnd().SetWindowText("&Save")
		// }?;

		Ok(())
	}
}
