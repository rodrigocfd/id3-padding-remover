use std::rc::Rc;
use winsafe::{self as w, co, shell};

use crate::ids::main as id;
use crate::util;
use crate::wnd_modify::WndModify;
use super::WndMain;

impl WndMain {
	pub(super) fn menu_events(&self) {
		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_OPEN, {
			let self2 = self.clone();
			move || {
				let fileo = w::CoCreateInstance::<shell::IFileOpenDialog>(
					&shell::clsid::FileOpenDialog,
					None,
					co::CLSCTX::INPROC_SERVER,
				).unwrap();

				fileo.SetOptions(
					fileo.GetOptions().unwrap()
						| shell::co::FOS::FORCEFILESYSTEM
						| shell::co::FOS::FILEMUSTEXIST
						| shell::co::FOS::ALLOWMULTISELECT,
				).unwrap();

				fileo.SetFileTypes(&[
					("MP3 audio files", "*.mp3"),
					("All files", "*.*"),
				]).unwrap();

				fileo.SetFileTypeIndex(1).unwrap();

				if fileo.Show(self2.wnd.hwnd()).unwrap() {
					self2.add_files(
						&fileo.GetResults().unwrap()
							.GetDisplayNames(shell::co::SIGDN::FILESYSPATH).unwrap(),
					).unwrap();
				}
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_EXCSEL, {
			let lst_files = self.lst_files.clone();
			move || {
				lst_files.items().delete(
					&lst_files.items().selected()).unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_MODIFY, {
			let self2 = self.clone();
			move || {
				let sel_files = self2.lst_files.columns().selected_texts(0);

				if sel_files.is_empty() {
					util::prompt::err(self2.wnd.hwnd(), "No files",
						"There are no selected files to be modified.");
					return;
				}

				let wa = WndModify::new(&self2.wnd, self2.tags_cache.clone(), Rc::new(sel_files));
				wa.show();

				{
					let tags_cache = self2.tags_cache.borrow();
					for idx in self2.lst_files.items().selected().iter() {
						let sel_file = self2.lst_files.items().text(*idx, 0);
						let tag = tags_cache.get(&sel_file).unwrap();
						self2.lst_files.items().set_text(*idx, 1, // write new padding
							&format!("{}", tag.original_padding())).unwrap();
					}
				}

				self2.show_selected_tag_frames().unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_CLR_DIACR, {
			let self2 = self.clone();
			move || {
				let clock = util::Timer::start();
				let sel_idxs = self2.lst_files.items().selected();

				{
					let mut tags_cache = self2.tags_cache.borrow_mut();

					for idx in sel_idxs.iter() {
						let file = self2.lst_files.items().text(*idx, 0);
						let file_new = util::clear_diacritics(&file);

						let tag = tags_cache.remove(&file).unwrap();
						tags_cache.insert(file_new.clone(), tag);
					}
				}

				for idx in sel_idxs.iter() {
					let file = self2.lst_files.items().text(*idx, 0);
					let file_new = util::clear_diacritics(&file);

					// This triggers LVN_ITEMCHANGED, which will borrow tags_cache.
					self2.lst_files.items().set_text(*idx, 0, &file_new).unwrap();
					w::MoveFile(&file, &file_new).unwrap();
				}

				util::prompt::info(self2.wnd.hwnd(), "Operation successful",
					&format!("Diacritics removed from {} file name(s) in {:.2} ms.",
						sel_idxs.len(), clock.now_ms()));
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_ABOUT, {
			let self2 = self.clone();
			move || {
				// Read version from resource.
				let exe_name = w::HINSTANCE::NULL.GetModuleFileName().unwrap();
				let mut res_buf = Vec::default();
				w::GetFileVersionInfo(&exe_name, &mut res_buf).unwrap();

				let vsffi = unsafe { w::VarQueryValue::<w::VS_FIXEDFILEINFO>(&res_buf, "\\").unwrap() };
				let ver = vsffi.dwFileVersion();

				util::prompt::info(self2.wnd.hwnd(), "About",
					&format!(
						"ID3 Padding Remover v{}.{}.{}\n\
						Writen in Rust with WinSafe library.\n\n\
						Rodrigo César de Freitas Dias © 2021",
						ver[0], ver[1], ver[2]));
			}
		});
	}
}
