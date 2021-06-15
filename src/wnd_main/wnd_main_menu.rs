use winsafe::{self as w, co, shell};

use crate::id3v2::{clear_diacritics, FrameData};
use crate::ids;
use super::WndMain;

impl WndMain {
	pub(super) fn menu_events(&self) {
		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_OPEN, {
			let self2 = self.clone();
			move || {
				let fileo: shell::IFileOpenDialog = w::CoCreateInstance(
					&shell::clsid::FileOpenDialog,
					None,
					co::CLSCTX::INPROC_SERVER,
				).unwrap();

				fileo.SetOptions(
					fileo.GetOptions().unwrap()
						| co::FOS::FORCEFILESYSTEM | co::FOS::FILEMUSTEXIST | co::FOS::ALLOWMULTISELECT,
				).unwrap();

				fileo.SetFileTypes(&[
					("MP3 audio files", "*.mp3"),
					("All files", "*.*"),
				]).unwrap();

				fileo.SetFileTypeIndex(0).unwrap();

				if fileo.Show(self2.wnd.hwnd()).unwrap() {
					self2.add_files(
						&fileo.GetResults().unwrap()
							.GetDisplayNames(co::SIGDN::FILESYSPATH).unwrap(),
					).unwrap();
				}
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_EXCSEL, {
			let lst_files = self.lst_files.clone();
			move || {
				lst_files.items().delete(
					&lst_files.items().selected()).unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_REMPAD, {
			let self2 = self.clone();
			move || {
				let freq = w::QueryPerformanceFrequency().unwrap();
				let t0 = w::QueryPerformanceCounter().unwrap();

				self2.write_selected_tags().unwrap(); // simply writing will remove padding

				self2.msg_info("Operation successful",
					&format!("Padding removed from {} file(s) in {:.2} ms.",
						self2.lst_files.items().selected_count(),
						((w::QueryPerformanceCounter().unwrap() - t0) as f64 / freq as f64) * 1000.0));
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_REMART, {
			let self2 = self.clone();
			move || {
				let freq = w::QueryPerformanceFrequency().unwrap();
				let t0 = w::QueryPerformanceCounter().unwrap();

				{
					let mut tags_cache = self2.tags_cache.borrow_mut();

					for file in self2.lst_files.columns().selected_texts(0).iter() {
						let tag = tags_cache.get_mut(file).unwrap();
						tag.frames_mut().retain(|f| f.name4() != "APIC");
					}
				}
				self2.write_selected_tags().unwrap();

				self2.msg_info("Operation successful",
					&format!("Album art removed from {} file(s) in {:.2} ms.",
						self2.lst_files.items().selected_count(),
						((w::QueryPerformanceCounter().unwrap() - t0) as f64 / freq as f64) * 1000.0));
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_REMRG, {
			let self2 = self.clone();
			move || {
				let freq = w::QueryPerformanceFrequency().unwrap();
				let t0 = w::QueryPerformanceCounter().unwrap();

				{
					let mut tags_cache = self2.tags_cache.borrow_mut();

					for file in self2.lst_files.columns().selected_texts(0).iter() {
						let tag = tags_cache.get_mut(file).unwrap();
						tag.frames_mut().retain(|f| {
							if f.name4() == "TXXX" {
								if let FrameData::MultiText(texts) = f.data() {
									if texts[0].starts_with("replaygain_") {
										return false;
									}
								}
							}
							true
						});
					}
				}
				self2.write_selected_tags().unwrap();

				self2.msg_info("Operation successful",
					&format!("ReplayGain removed from {} file(s) in {:.2} ms.",
						self2.lst_files.items().selected_count(),
						((w::QueryPerformanceCounter().unwrap() - t0) as f64 / freq as f64) * 1000.0));
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_PRXYEAR, {
			let self2 = self.clone();
			move || {
				let freq = w::QueryPerformanceFrequency().unwrap();
				let t0 = w::QueryPerformanceCounter().unwrap();

				{
					let mut tags_cache = self2.tags_cache.borrow_mut();

					for file in self2.lst_files.columns().selected_texts(0).iter() {
						let tag = tags_cache.get_mut(file).unwrap();
						let frames = tag.frames_mut();

						let year = if let Some(year_frame) = frames.iter().find(|f| f.name4() == "TYER") {
							if let FrameData::Text(text) = year_frame.data() {
								text.clone()
							} else {
								self2.msg_err("Bad frame",
									&format!("File: {}\n\nYear frame has the wrong data type.", file));
								return
							}
						} else {
							self2.msg_err("Missing frame",
								&format!("File: {}\n\nYear frame not found.", file));
							return
						};

						let album = if let Some(album_frame) = frames.iter_mut().find(|f| f.name4() == "TALB") {
							if let FrameData::Text(text) = album_frame.data_mut() {
								text
							} else {
								self2.msg_err("Bad frame",
									&format!("File: {}\n\nAlbum frame has the wrong data type.", file));
								return
							}
						} else {
							self2.msg_err("Missing frame",
								&format!("File: {}\n\nAlbum frame not found.", file));
							return
						};

						if album.starts_with(&year) {
							let res = self2.wnd.hwnd().TaskDialog(
								None,
								Some(ids::TITLE),
								Some("Dubious data"),
								Some(&format!("File: {}\n\n\
									Album appears to have the year prefix {}.\n\
									Continue anyway?", file, year)),
								co::TDCBF::OK | co::TDCBF::CANCEL,
								w::IdTdicon::Tdicon(co::TD_ICON::WARNING),
							).unwrap();
							if res != co::DLGID::OK {
								return;
							}
						}
						*album = format!("{} {}", year, album);
					}
				}
				self2.write_selected_tags().unwrap();

				self2.msg_info("Operation successful",
					&format!("Prefix saved in {} file(s) in {:.2} ms.",
						self2.lst_files.items().selected_count(),
						((w::QueryPerformanceCounter().unwrap() - t0) as f64 / freq as f64) * 1000.0));
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_CLRDIAC, {
			let self2 = self.clone();
			move || {
				let freq = w::QueryPerformanceFrequency().unwrap();
				let t0 = w::QueryPerformanceCounter().unwrap();

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

				self2.msg_info("Operation successful",
					&format!("Diacritics removed from {} file name(s) in {:.2} ms.",
						sel_idxs.len(),
						((w::QueryPerformanceCounter().unwrap() - t0) as f64 / freq as f64) * 1000.0));
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_ABOUT, {
			let self2 = self.clone();
			move || {
				// Read version from resource.
				let exe_name = w::HINSTANCE::NULL.GetModuleFileName().unwrap();
				let mut res_buf = Vec::default();
				w::GetFileVersionInfo(&exe_name, &mut res_buf).unwrap();

				let fis = w::VarQueryValue(&res_buf, "\\").unwrap();
				let fi: &w::VS_FIXEDFILEINFO = unsafe { &*(fis.as_ptr() as *const w::VS_FIXEDFILEINFO) };
				let ver = fi.dwFileVersion();

				self2.msg_info("About", &format!(
					"ID3 Padding Remover v{}.{}.{}\n\
					Writen in Rust with WinSafe library.\n\n\
					Rodrigo César de Freitas Dias © 2021",
					ver[0], ver[1], ver[2]));
			}
		});
	}
}
