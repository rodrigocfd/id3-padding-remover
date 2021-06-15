use winsafe::{self as w, co, msg};

use crate::ids;
use super::PreDelete;
use super::WndMain;

impl WndMain {
	pub(super) fn events(&self) {
		self.wnd.on().wm_init_dialog({
			let self2 = self.clone();
			move |_: msg::wm::InitDialog| -> bool {
				// Files list view.
				self2.lst_frames.toggle_extended_style(true, co::LVS_EX::GRIDLINES);

				// Since it doesn't have LVS_SHAREIMAGELISTS style, the image list
				// will be automatically deleted by the list view.
				let himgl = w::HIMAGELIST::Create(16, 16, co::ILC::COLOR32, 1, 1).unwrap();
				himgl.AddIconFromShell(&["mp3"]).unwrap();
				self2.lst_files.set_image_list(co::LVSIL::SMALL, himgl);

				self2.lst_files.columns().add(&[
					("File", 0),
					("Padding", 60),
				]).unwrap();
				self2.lst_files.columns().set_width_to_fill(0).unwrap();

				// Frames list view.
				self2.lst_frames.columns().add(&[
					("Frame", 65),
					("Value", 0),
				]).unwrap();
				self2.lst_frames.columns().set_width_to_fill(1).unwrap();
				self2.lst_frames.hwnd().EnableWindow(false);

				self2.titlebar_count(PreDelete::No).unwrap();
				true
			}
		});

		self.wnd.on().wm_size({
			let self2 = self.clone();
			move |p: msg::wm::Size| {
				if p.request == co::SIZE_R::MINIMIZED {
					return;
				}

				self2.lst_files.columns().set_width_to_fill(0).unwrap();
				self2.lst_frames.columns().set_width_to_fill(1).unwrap();
			}
		});

		self.wnd.on().wm_init_menu_popup({
			let self2 = self.clone();
			move |p: msg::wm::InitMenuPopup| {
				if p.hmenu == self2.lst_files.context_menu().unwrap() {
					let has_sel = self2.lst_files.items().selected_count() > 0;

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
			let self2 = self.clone();
			move |p: msg::wm::DropFiles| {
				let dropped_files = p.hdrop.DragQueryFiles().unwrap();
				let mut all_files = Vec::with_capacity(dropped_files.len());

				for mut file in dropped_files.into_iter() {
					if w::GetFileAttributes(&file).unwrap().has(co::FILE_ATTRIBUTE::DIRECTORY) {
						if !file.ends_with('\\') {
							file.push('\\');
						}
						file.push_str("*.mp3");

						for mp3 in w::HFINDFILE::ListAll(&file).unwrap() { // just search 1 level below
							if mp3.to_lowercase().ends_with(".mp3") {
								all_files.push(mp3);
							}
						}
					} else if file.to_lowercase().ends_with(".mp3") {
						all_files.push(file);
					}
				}

				self2.add_files(&all_files).unwrap();
			}
		});

		self.lst_files.on().lvn_key_down({
			let self2 = self.clone();
			move |p: &w::NMLVKEYDOWN| {
				if p.wVKey == co::VK::DELETE { // delete item on DEL
					self2.wnd.hwnd().SendMessage(msg::wm::Command {
						event: w::AccelMenuCtrl::Menu(ids::MNU_FILE_EXCSEL as _),
					});
				}
			}
		});

		self.lst_files.on().lvn_item_changed({
			let self2 = self.clone();
			move |_: &w::NMLISTVIEW| {
				self2.show_tag_frames().unwrap();
				self2.titlebar_count(PreDelete::No).unwrap();
			}
		});

		self.lst_files.on().lvn_delete_item({
			let self2 = self.clone();
			move |p: &w::NMLISTVIEW| {
				self2.tags_cache.borrow_mut() // remove entry from cache
					.remove(&self2.lst_files.items().text_str(p.iItem as _, 0));
				self2.titlebar_count(PreDelete::Yes).unwrap();
			}
		});
	}
}
