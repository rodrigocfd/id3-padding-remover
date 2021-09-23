use winsafe::{self as w, co, msg};

use crate::ids::main as id;
use super::PreDelete;
use super::WndMain;

impl WndMain {
	pub(super) fn events(&self) {
		self.wnd.on().wm_init_dialog({
			let self2 = self.clone();
			move |_| {
				// Files list view.
				self2.lst_frames.toggle_extended_style(true, co::LVS_EX::GRIDLINES);

				// Since it doesn't have LVS_SHAREIMAGELISTS style, the image list
				// will be automatically deleted by the list view.
				let himgl = w::HIMAGELIST::Create(w::SIZE::new(16, 16), co::ILC::COLOR32, 1, 1)?;
				himgl.AddIconFromShell(&["mp3"])?;
				self2.lst_files.set_image_list(co::LVSIL::SMALL, himgl);

				self2.lst_files.columns().add(&[
					("File", 0),
					("Padding", 60),
				])?;
				self2.lst_files.columns().set_width_to_fill(0)?;

				// Frames list view.
				self2.lst_frames.columns().add(&[
					("Frame", 65),
					("Value", 0),
				])?;
				self2.lst_frames.columns().set_width_to_fill(1)?;
				self2.lst_frames.hwnd().EnableWindow(false);

				self2.titlebar_count(PreDelete::No)?;
				Ok(true)
			}
		});

		self.wnd.on().wm_size({
			let self2 = self.clone();
			move |p| {
				if p.request == co::SIZE_R::MINIMIZED {
					return Ok(());
				}

				self2.lst_files.columns().set_width_to_fill(0)?;
				self2.lst_frames.columns().set_width_to_fill(1)?;
				Ok(())
			}
		});

		self.wnd.on().wm_init_menu_popup({
			let self2 = self.clone();
			move |p| {
				if p.hmenu == self2.lst_files.context_menu().unwrap() {
					let has_sel = self2.lst_files.items().selected_count() > 0;

					[id::MNU_FILE_EXCSEL, id::MNU_FILE_MODIFY, id::MNU_FILE_CLR_DIACR]
						.iter().for_each(|id| {
							p.hmenu.EnableMenuItem(w::IdPos::Id(*id), has_sel).unwrap(); // FIXME
						});
				}
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(co::DLGID::CANCEL.into(), {
			let wnd = self.wnd.clone();
			move || {
				wnd.hwnd().SendMessage(msg::wm::Close {}); // close on ESC
				Ok(())
			}
		});

		self.wnd.on().wm_drop_files({
			let self2 = self.clone();
			move |p| {
				let dropped_files = p.hdrop.DragQueryFiles()?;
				let mut all_files = Vec::with_capacity(dropped_files.len());

				for mut file in dropped_files.into_iter() {
					if w::GetFileAttributes(&file)?.has(co::FILE_ATTRIBUTE::DIRECTORY) {
						if !file.ends_with('\\') {
							file.push('\\');
						}
						file.push_str("*.mp3");

						for mp3 in w::HFINDFILE::ListAll(&file)? { // just search 1 level below
							if mp3.to_lowercase().ends_with(".mp3") {
								all_files.push(mp3);
							}
						}
					} else if file.to_lowercase().ends_with(".mp3") {
						all_files.push(file);
					}
				}

				self2.add_files(&all_files)?;
				Ok(())
			}
		});

		self.lst_files.on().lvn_key_down({
			let self2 = self.clone();
			move |p| {
				if p.wVKey == co::VK::DELETE { // delete item on DEL
					self2.wnd.hwnd().SendMessage(msg::wm::Command {
						event: w::AccelMenuCtrl::Menu(id::MNU_FILE_EXCSEL as _),
					});
				}
				Ok(())
			}
		});

		self.lst_files.on().lvn_item_changed({
			let self2 = self.clone();
			move |_| {
				self2.show_selected_tag_frames()?;
				self2.titlebar_count(PreDelete::No)?;
				Ok(())
			}
		});

		self.lst_files.on().lvn_delete_item({
			let self2 = self.clone();
			move |p| {
				self2.tags_cache.borrow_mut() // remove entry from cache
					.remove(&self2.lst_files.items().text(p.iItem as _, 0));
				self2.titlebar_count(PreDelete::Yes)?;
				Ok(())
			}
		});
	}
}
