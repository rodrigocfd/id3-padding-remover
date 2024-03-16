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

		self.field_packs.borrow()
			.iter()
			.for_each(|field_pack| {
				let fp2 = field_pack.clone();
				field_pack.chk.on().bn_clicked(move || {
					if fp2.chk.is_checked() {
						fp2.txt.hwnd().EnableWindow(true);
						fp2.txt.focus();
					} else {
						fp2.txt.hwnd().EnableWindow(false);
					}
					Ok(())
				});
			});
	}
}
