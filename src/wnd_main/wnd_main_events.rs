use winsafe::co;
use winsafe::msg;

use super::WndMain;

impl WndMain {
	pub(super) fn events(&self) {
		self.wnd.on().wm_init_dialog({
			let selfc = self.clone();
			move |_: msg::wm::InitDialog| -> bool {
				selfc.lst_files.columns().add(&[
					("File", 0),
					("Padding", 60),
				]).unwrap();
				selfc.lst_files.columns().set_width_to_fill(0).unwrap();

				selfc.lst_frames.columns().add(&[
					("Frame", 65),
					("Value", 0),
				]).unwrap();
				selfc.lst_frames.columns().set_width_to_fill(1).unwrap();

				true
			}
		});

		self.wnd.on().wm_size({
			let selfc = self.clone();
			move |p: msg::wm::Size| {
				if p.request == co::SIZE_R::MINIMIZED {
					return;
				}

				selfc.lst_files.columns().set_width_to_fill(0).unwrap();
				selfc.lst_frames.columns().set_width_to_fill(1).unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(co::DLGID::CANCEL.into(), { // close on ESC
			let wnd = self.wnd.clone();
			move || {
				wnd.hwnd().PostMessage(msg::wm::Close {}).unwrap();
			}
		});
	}
}
