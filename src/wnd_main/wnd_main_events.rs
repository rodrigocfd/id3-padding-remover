use winsafe as w;
use winsafe::co;
use winsafe::msg;

use crate::ids;
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

		self.wnd.on().wm_init_menu_popup({
			let selfc = self.clone();
			move |p: msg::wm::InitMenuPopup| {
				if p.hmenu == selfc.lst_files.context_menu().unwrap() {
					let has_sel = selfc.lst_files.items().selected_count() > 0;
					p.hmenu.EnableMenuItem(w::IdPos::Id(ids::MNU_FILE_EXCSEL), has_sel).unwrap();
					p.hmenu.EnableMenuItem(w::IdPos::Id(ids::MNU_FILE_REMPAD), has_sel).unwrap();
				}
			}
		});

		self.wnd.on().wm_command_accel_menu(co::DLGID::CANCEL.into(), { // close on ESC
			let wnd = self.wnd.clone();
			move || {
				wnd.hwnd().PostMessage(msg::wm::Close {}).unwrap();
			}
		});

		self.lst_files.on().lvn_key_down({
			let selfc = self.clone();
			move |p: &w::NMLVKEYDOWN| {
				if p.wVKey == co::VK::DELETE { // delete item on DEL
					selfc.wnd.hwnd().SendMessage(msg::wm::Command {
						event: w::AccelMenuCtrl::Menu(ids::MNU_FILE_EXCSEL as _),
					});
				}
			}
		});

		self.lst_files.on().lvn_item_changed({
			let selfc = self.clone();
			move |_: &w::NMLISTVIEW| {
				selfc.lst_frames.items().delete_all().unwrap();
				let sel_files = selfc.lst_files.columns().selected_texts(0);
				if sel_files.len() == 0 {
					return;
				}

				if sel_files.len() == 1 {
					let tags = selfc.tags_cache.borrow();
					let tag = tags.get(&sel_files[0]).unwrap();
					selfc.show_tag_frames(&tag).unwrap();

				} else { // multiple selected items, just display a placeholder
					selfc.lst_frames.items().add("", None).unwrap();
					selfc.lst_frames.items().set_text(0, 1,
						&format!("{} selected...", sel_files.len())).unwrap();
				}
			}
		});
	}
}
