use winsafe::{self as w, prelude::*, co, gui, msg};

use crate::ids;
use super::WndMain;

impl WndMain {
	pub(super) fn context_menu(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_FILE_OPEN, move || {

			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_FILE_EXIT, move || {
			self2.wnd.hwnd().SendMessage(msg::wm::Close {});
			Ok(())
		});
	}
}
