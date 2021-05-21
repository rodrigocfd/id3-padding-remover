use winsafe as w;
use winsafe::co;
use winsafe::msg;

use super::WndMain;

impl WndMain {
	pub(super) fn events(&self) {
		self.wnd.on().wm_init_dialog({
			let selfc = self.clone();
			move |_: msg::wm::InitDialog| -> bool {
				selfc.lst_frames.toggle_extended_style(true, co::LVS_EX::GRIDLINES);

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

		self.lst_files.on().lvn_item_changed({
			let selfc = self.clone();
			move |_: &w::NMLISTVIEW| {
				selfc.lst_frames.items().delete_all().unwrap();

				let first_sel_idx = selfc.lst_files.items().selected()[0];
				let file = selfc.lst_files.items().text_str(first_sel_idx, 0);

				let tags = selfc.tags.borrow();
				let tag = tags.get(&file).unwrap();

				selfc.show_tag_frames(&tag);
			}
		});
	}
}
