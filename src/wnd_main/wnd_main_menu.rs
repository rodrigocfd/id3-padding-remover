use winsafe as w;
use winsafe::co;
use winsafe::shell;

use crate::id3v2::FrameData;
use crate::ids;
use super::WndMain;

impl WndMain {
	pub(super) fn menu_events(&self) {
		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_OPEN, {
			let selfc = self.clone();
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

				if fileo.Show(selfc.wnd.hwnd()).unwrap() {
					selfc.add_files(
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
			let selfc = self.clone();
			move || {
				selfc.write_selected_tags().unwrap(); // simply writing will remove padding
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_REMART, {
			let selfc = self.clone();
			move || {
				{
					let mut tags_cache = selfc.tags_cache.borrow_mut();

					for file in selfc.lst_files.columns().selected_texts(0).iter() {
						let tag = tags_cache.get_mut(file).unwrap();
						tag.frames_mut().retain(|f| f.name4() != "APIC");
					}
				}
				selfc.write_selected_tags().unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_REMRG, {
			let selfc = self.clone();
			move || {
				{
					let mut tags_cache = selfc.tags_cache.borrow_mut();

					for file in selfc.lst_files.columns().selected_texts(0).iter() {
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
				selfc.write_selected_tags().unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_PRXYEAR, {
			let selfc = self.clone();
			move || {
				{
					let mut tags_cache = selfc.tags_cache.borrow_mut();

					for file in selfc.lst_files.columns().selected_texts(0).iter() {
						let tag = tags_cache.get_mut(file).unwrap();
						let frames = tag.frames_mut();

						let year = match frames.iter().find(|f| f.name4() == "TYER") {
							None => {
								selfc.wnd.hwnd().MessageBox(
									"Year frame not found.",
									"No frame", co::MB::ICONEXCLAMATION).unwrap();
								return
							},
							Some(year_frame) => match year_frame.data() {
								FrameData::Text(text) => text.clone(),
								_ => {
									selfc.wnd.hwnd().MessageBox(
										"Year frame has the wrong data type.",
										"Bad frame", co::MB::ICONEXCLAMATION).unwrap();
									return
								},
							},
						};

						match frames.iter_mut().find(|f| f.name4() == "TALB") {
							None => {
								selfc.wnd.hwnd().MessageBox(
									"Album frame not found.",
									"No frame", co::MB::ICONEXCLAMATION).unwrap();
								return
							},
							Some(album_frame) => match album_frame.data_mut() {
								FrameData::Text(text) => {
									if text.starts_with(&year) {
										let res = selfc.wnd.hwnd().MessageBox(
											&format!("Album appears to have the year prefix: {}.\nContinue?", year),
											"Verify action",
											co::MB::ICONEXCLAMATION | co::MB::YESNO).unwrap();
										if res != co::DLGID::YES {
											return;
										}
									}
									*text = format!("{} {}", year, text);
								},
								_ => {
									selfc.wnd.hwnd().MessageBox(
										"Album frame has the wrong data type.",
										"Bad frame", co::MB::ICONEXCLAMATION).unwrap();
									return
								},
							},
						}
					}
				}
				selfc.write_selected_tags().unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_SIMPLEN, {
			let selfc = self.clone();
			move || {

			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_ABOUT, {
			let wnd = self.wnd.clone();
			move || {
				wnd.hwnd().MessageBox(
					"ID3 Padding Remover v2\n\
					Writen in Rust with WinSafe library.\n\n\
					Rodrigo César de Freitas Dias © 2021",
					"About",
					co::MB::ICONINFORMATION,
				).unwrap();
			}
		});
	}
}