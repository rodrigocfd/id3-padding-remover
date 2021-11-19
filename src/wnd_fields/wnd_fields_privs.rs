use winsafe::prelude::*;

use super::WndFields;

impl WndFields {
	pub(super) fn _enable_buttons_if_at_least_one_checked(&self) {
		let at_least_one_checked = self.fields.iter()
			.find(|field| field.chk.is_checked())
			.is_some();

		self.btn_clear_checks.hwnd().EnableWindow(at_least_one_checked);
		self.btn_save.hwnd().EnableWindow(at_least_one_checked);
	}
}
