use winsafe::{prelude::*, self as w, co, msg};

use crate::util;
use super::{ids, PreDelete, TagOp, WndMain};

impl WndMain {
	pub(super) fn _events(&self) -> w::ErrResult<()> {
		self.wnd.on().wm_init_dialog({
			let self2 = self.clone();
			move |_| {
				// Since it doesn't have LVS_SHAREIMAGELISTS style, the image list
				// will be automatically deleted by the list view.
				let himgl = w::HIMAGELIST::Create(w::SIZE::new(16, 16), co::ILC::COLOR32, 1, 1)?;
				himgl.AddIconFromShell(&["mp3"])?;
				self2.lst_mp3s.set_image_list(co::LVSIL::SMALL, himgl);

				self2.lst_mp3s.columns().add(&[
					("File", 0),
					("Padding", 60),
				])?;
				self2.lst_mp3s.columns().set_width_to_fill(0)?;

				self2.lst_frames.set_extended_style(true, co::LVS_EX::GRIDLINES);
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
				if p.request != co::SIZE_R::MINIMIZED {
					self2.lst_mp3s.columns().set_width_to_fill(0)?;
					self2.lst_frames.columns().set_width_to_fill(1)?;
				}
				Ok(())
			}
		});

		self.wnd.on().wm_init_menu_popup({
			let self2 = self.clone();
			move |p| {
				if p.hmenu == self2.lst_mp3s.context_menu().unwrap() {
					let sel_count = self2.lst_mp3s.items().selected_count() > 0;

					[ids::MNU_MP3S_DELETE,
						ids::MNU_MP3S_REM_PAD, ids::MNU_MP3S_REM_RG, ids::MNU_MP3S_REM_RG_PIC,
						ids::MNU_MP3S_DEL_TAG,
						ids::MNU_MP3S_COPY_TO_FOLDER, ids::MNU_MP3S_RENAME, ids::MNU_MP3S_RENAME_PREFIX,
					].iter()
						.map(|id| p.hmenu.EnableMenuItem(w::IdPos::Id(*id), sel_count))
						.collect::<w::WinResult<Vec<_>>>()?;

				} else if p.hmenu == self2.lst_frames.context_menu().unwrap() {
					p.hmenu.EnableMenuItem(
						w::IdPos::Id(ids::MNU_FRAMES_REM),
						self2.lst_frames.items()
							.iter_selected()
							.find(|sel_item| !sel_item.text(0).is_empty())
							.is_some(),
					)?;

					p.hmenu.EnableMenuItem(
						w::IdPos::Id(ids::MNU_FRAMES_MOVE_UP),
						self2.lst_frames.items().selected_count() == 1
							&& self2.lst_frames.items()
								.iter()
								.enumerate()
								.find(|(idx, item)| *idx != 0
									&& item.is_selected()
									&& !item.text(0).is_empty())
								.is_some(),
					)?;
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
				let mut all_files = Vec::with_capacity(5); // arbitrary

				for file in p.hdrop.iter() {
					let mut file = file?;
					if w::GetFileAttributes(&file)?.has(co::FILE_ATTRIBUTE::DIRECTORY) {
						if !file.ends_with('\\') { file.push('\\'); }
						file.push_str("*.mp3");

						for mp3 in w::HFINDFILE::iter(&file) { // search just 1 level below
							let mp3 = mp3?;
							if w::path::has_extension(&mp3, &[".mp3"]) {
								all_files.push(mp3);
							}
						}
					} else if w::path::has_extension(&file, &[".mp3"]) {
						all_files.push(file);
					}
				}

				if let Err(e) = self2._modal_tag_op(TagOp::Load, &all_files) {
					util::prompt::err(self2.wnd.hwnd(),
						"Error", Some("Tag load failed"), &e.to_string())?;
				}

				Ok(())
			}
		});

		self.lst_mp3s.on().lvn_key_down({
			let self2 = self.clone();
			move |p| {
				if p.wVKey == co::VK::DELETE { // delete item on DEL
					self2.wnd.hwnd().SendMessage(msg::wm::Command { // simulate menu click
						event: w::AccelMenuCtrl::Menu(ids::MNU_MP3S_DELETE as _),
					});
				}
				Ok(())
			}
		});

		self.lst_mp3s.on().lvn_item_changed({
			let self2 = self.clone();
			move |_| {
				self2._display_sel_tags_frames()?;

				self2.wnd_fields.feed(
					self2.lst_mp3s.items()
						.iter_selected()
						.map(|sel_item| sel_item.text(0))
						.collect::<Vec<_>>(),
				)?;

				self2._titlebar_count(PreDelete::No)?;
				Ok(())
			}
		});

		self.lst_mp3s.on().lvn_delete_item({
			let self2 = self.clone();
			move |p| {
				let file_path = self2.lst_mp3s.items().get(p.iItem as _).text(0);
				self2.tags_cache.lock().unwrap().remove(&file_path); // remove entry from cache
				self2._titlebar_count(PreDelete::Yes)?;
				Ok(())
			}
		});

		self.wnd_fields.on_save({
			let self2 = self.clone();
			move || {
				let clock = util::Timer::start()?;

				if let Err(e) = self2._modal_tag_op( // WndFields won't save the tags
					TagOp::SaveAndLoad,
					&self2.lst_mp3s.items()
						.iter_selected()
						.map(|sel_item| sel_item.text(0))
						.collect::<Vec<_>>(),
				) {
					util::prompt::err(self2.wnd.hwnd(),
						"Error", Some("Tag updating failed"), &e.to_string())?;
				} else {
					self2._display_sel_tags_frames()?;
					util::prompt::info(self2.wnd.hwnd(),
						"Operation successful", Some("Success"),
						&format!("Tag updated in {} file(s) in {:.2} ms.",
							self2.lst_mp3s.items().selected_count(), clock.now_ms()?))?;
				}

				self2.lst_mp3s.focus()?;
				Ok(())
			}
		})?;

		Ok(())
	}
}
