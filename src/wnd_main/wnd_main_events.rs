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

				selfc.titlebar_count(false).unwrap();
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

					[ids::MNU_FILE_EXCSEL, ids::MNU_FILE_REMPAD, ids::MNU_FILE_REMART,
						ids::MNU_FILE_REMRG, ids::MNU_FILE_PRXYEAR, ids::MNU_FILE_CLRDIAC,
					].iter()
						.for_each(|id| {
							p.hmenu.EnableMenuItem(w::IdPos::Id(*id), has_sel).unwrap();
						});
				}
			}
		});

		self.wnd.on().wm_command_accel_menu(co::DLGID::CANCEL.into(), { // close on ESC
			let wnd = self.wnd.clone();
			move || {
				wnd.hwnd().PostMessage(msg::wm::Close {}).unwrap();
			}
		});

		self.wnd.on().wm_drop_files({
			let selfc = self.clone();
			move |p: msg::wm::DropFiles| {
				selfc.add_files(
					&p.hdrop.DragQueryFiles().unwrap(),
				).unwrap();
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
				selfc.show_tag_frames().unwrap();
				selfc.titlebar_count(false).unwrap();
			}
		});

		self.lst_files.on().lvn_delete_item({
			let selfc = self.clone();
			move |p: &w::NMLISTVIEW| {
				selfc.tags_cache.borrow_mut() // remove entry from cache
					.remove(&selfc.lst_files.items().text_str(p.iItem as _, 0));
				selfc.titlebar_count(true).unwrap();
			}
		});
	}
}
