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

			}
		});
	}
}
