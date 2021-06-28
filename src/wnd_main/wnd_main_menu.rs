use std::rc::Rc;
use winsafe::{self as w, co, shell};

use crate::id3v2::clear_diacritics;
use crate::ids::{APP_TITLE, main as id};
use crate::util;
use crate::wnd_modify::WndModify;
use super::WndMain;

impl WndMain {
	pub(super) fn menu_events(&self) {
		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_OPEN, {
			let self2 = self.clone();
			move || {
				let fileo: shell::IFileOpenDialog = w::CoCreateInstance(
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

				fileo.SetFileTypeIndex(0).unwrap();

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
					self2.wnd.hwnd().TaskDialog(None, Some(APP_TITLE),
						Some("No files"),
						Some("There are no selected files to be modified."),
						co::TDCBF::OK, w::IdTdicon::Tdicon(co::TD_ICON::ERROR)).unwrap();
					return;
				}

				let wa = WndModify::new(&self2.wnd, self2.tags_cache.clone(), Rc::new(sel_files));
				wa.show();

				{
					let tags_cache = self2.tags_cache.borrow();
					let mut buf = w::WString::default();
					for i in 0..self2.lst_files.items().count() {
						self2.lst_files.items().text(i, 0, &mut buf);
						let tag = tags_cache.get(&buf.to_string()).unwrap();

						self2.lst_files.items().set_text(i, 1, // write new padding
							&format!("{}", tag.original_padding())).unwrap();
					}
				}

				self2.show_selected_tag_frames().unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_CLR_DIACR, {
			let self2 = self.clone();
			move || {
				let t0 = util::timer_start();
				let sel_idxs = self2.lst_files.items().selected();

				{
					let mut tags_cache = self2.tags_cache.borrow_mut();

					for idx in sel_idxs.iter() {
						let file = self2.lst_files.items().text_str(*idx, 0);
						let file_new = clear_diacritics(&file);

						let tag = tags_cache.remove(&file).unwrap();
						tags_cache.insert(file_new.clone(), tag);
					}
				}

				for idx in sel_idxs.iter() {
					let file = self2.lst_files.items().text_str(*idx, 0);
					let file_new = clear_diacritics(&file);

					// This triggers LVN_ITEMCHANGED, which will borrow tags_cache.
					self2.lst_files.items().set_text(*idx, 0, &file_new).unwrap();
					w::MoveFile(&file, &file_new).unwrap();
				}

				self2.wnd.hwnd().TaskDialog(None, Some(APP_TITLE),
					Some("Operation successful"),
					Some(&format!("Diacritics removed from {} file name(s) in {:.2} ms.",
						sel_idxs.len(), util::timer_end_ms(t0))),
					co::TDCBF::OK,
					w::IdTdicon::Tdicon(co::TD_ICON::INFORMATION)).unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_ABOUT, {
			let self2 = self.clone();
			move || {
				// Read version from resource.
				let exe_name = w::HINSTANCE::NULL.GetModuleFileName().unwrap();
				let mut res_buf = Vec::default();
				w::GetFileVersionInfo(&exe_name, &mut res_buf).unwrap();

				let fis = w::VarQueryValue(&res_buf, "\\").unwrap();
				let fi: &w::VS_FIXEDFILEINFO = unsafe { &*(fis.as_ptr() as *const w::VS_FIXEDFILEINFO) };
				let ver = fi.dwFileVersion();

				self2.wnd.hwnd().TaskDialog(None, Some(APP_TITLE),
					Some("About"),
					Some(&format!(
						"ID3 Padding Remover v{}.{}.{}\n\
						Writen in Rust with WinSafe library.\n\n\
						Rodrigo César de Freitas Dias © 2021",
						ver[0], ver[1], ver[2])),
					co::TDCBF::OK, w::IdTdicon::Tdicon(co::TD_ICON::INFORMATION)).unwrap();
			}
		});
	}
}
