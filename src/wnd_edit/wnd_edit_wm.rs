use winsafe::{self as w, prelude::*, co, gui, msg};

use super::WndEdit;

impl WndEdit {
	pub(super) fn wm_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.init_dialog()
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(co::DLGID::OK.into(), move || {

			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(co::DLGID::CANCEL.into(), move || {
			self2.wnd.hwnd().SendMessage(msg::wm::Close {}); // close on ESC
			Ok(())
		});
	}
}
