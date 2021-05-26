use winsafe as w;
use winsafe::co;
use winsafe::shell;

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

						if let Some(apic_idx) = tag.frames().iter().position(|f| f.name4() == "APIC") {
							tag.frames_mut().remove(apic_idx);
						}
					}
				}
				selfc.write_selected_tags().unwrap();
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_REMRG, {
			let selfc = self.clone();
			move || {

			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_FILE_PRXYEAR, {
			let selfc = self.clone();
			move || {

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
