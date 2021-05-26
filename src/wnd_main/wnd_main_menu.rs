use winsafe as w;
use winsafe::co;
use winsafe::shell;

use crate::id3v2::Tag;
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
				let mut tags_cache = selfc.tags_cache.borrow_mut();
				let sel_idxs = selfc.lst_files.items().selected();

				for idx in sel_idxs.iter() {
					let file = selfc.lst_files.items().text_str(*idx, 0);
					let tag = tags_cache.get_mut(&file).unwrap();
					tag.write(&file).unwrap(); // save tag to file, no padding is written

					*tag = Tag::read(&file).unwrap(); // load tag back from file
					selfc.lst_files.items().set_text(*idx, 1,
						&format!("{}", tag.original_padding())).unwrap(); // update padding info
				}

				selfc.show_tag_frames().unwrap();
			}
		});
	}
}
