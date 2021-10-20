use winsafe::{prelude::*, self as w, co, msg};

use crate::ids::main as id;
use super::PreDelete;
use super::WndMain;

impl WndMain {
	pub(super) fn _events(&self) {
		self.wnd.on().wm_init_dialog({
			let self2 = self.clone();
			move |_| {
				// Files list view.
				self2.lst_frames.set_extended_style(true, co::LVS_EX::GRIDLINES);

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

				self2._titlebar_count(PreDelete::No)?;
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

					[id::MNU_FILE_DELSEL, id::MNU_FILE_REMPAD,
						id::MNU_FILE_REMRG, id::MNU_FILE_REMRGART,
						id::MNU_FILE_RENAME, id::MNU_FILE_RENAMETRCK,
					].iter()
						.map(|id| p.hmenu.EnableMenuItem(w::IdPos::Id(*id), has_sel))
						.collect::<Result<Vec<_>, _>>()?;
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
				let mut all_files = Vec::default();

				for file in p.hdrop.iter() {
					let mut file = file?;
					if w::GetFileAttributes(&file)?.has(co::FILE_ATTRIBUTE::DIRECTORY) {
						if !file.ends_with('\\') {
							file.push('\\');
						}
						file.push_str("*.mp3");

						for mp3 in w::HFINDFILE::iter(&file) { // search just 1 level below
							let mp3 = mp3?;
							if mp3.to_lowercase().ends_with(".mp3") {
								all_files.push(mp3);
							}
						}
					} else if file.to_lowercase().ends_with(".mp3") {
						all_files.push(file);
					}
				}

				self2._add_files(&all_files)?;
				Ok(())
			}
		});

		self.lst_files.on().lvn_key_down({
			let self2 = self.clone();
			move |p| {
				if p.wVKey == co::VK::DELETE { // delete item on DEL
					self2.wnd.hwnd().SendMessage(msg::wm::Command { // simulate menu click
						event: w::AccelMenuCtrl::Menu(id::MNU_FILE_DELSEL as _),
					});
				}
				Ok(())
			}
		});

		self.lst_files.on().lvn_item_changed({
			let self2 = self.clone();
			move |_| {
				self2._display_sel_tags_frames()?;
				self2.wnd_fields.show_text_fields(
					self2.lst_files.items()
						.iter_selected()
						.map(|item| item.text(0))
						.collect::<Vec<_>>(),
				)?;
				self2._titlebar_count(PreDelete::No)?;
				Ok(())
			}
		});

		self.lst_files.on().lvn_delete_item({
			let self2 = self.clone();
			move |p| {
				let file_path = self2.lst_files.items().get(p.iItem as _).text(0);
				self2.tags_cache.try_borrow_mut()?.remove(&file_path); // remove entry from cache
				self2._titlebar_count(PreDelete::Yes)?;
				Ok(())
			}
		});

		self.wnd_fields.on_save({
			let self2 = self.clone();
			move || {
				self2._add_files( // reload all tags from their files
					&self2.lst_files.items().iter_selected()
						.map(|item| item.text(0))
						.collect::<Vec<_>>(),
				)
			}
		});
	}
}
