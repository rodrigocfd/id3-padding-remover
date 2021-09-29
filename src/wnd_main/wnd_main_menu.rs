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
				)?;

				fileo.SetOptions(
					fileo.GetOptions()?
						| shell::co::FOS::FORCEFILESYSTEM
						| shell::co::FOS::FILEMUSTEXIST
						| shell::co::FOS::ALLOWMULTISELECT,
				)?;

				fileo.SetFileTypes(&[
					("MP3 audio files", "*.mp3"),
					("All files", "*.*"),
				])?;

				fileo.SetFileTypeIndex(1)?;

				// let sh_dir = shell::IShellItem::from_path(&w::GetCurrentDirectory()?)?;
				// fileo.SetFolder(&sh_dir)?;

				if fileo.Show(self2.wnd.hwnd())? {
					self2.add_files(
						&fileo.GetResults()?
							.GetDisplayNames(shell::co::SIGDN::FILESYSPATH)?,
					)?;
				}
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_EXCSEL, {
			let lst_files = self.lst_files.clone();
			move || {
				lst_files.items().delete(
					&lst_files.items().selected())?;
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_MODIFY, {
			let self2 = self.clone();
			move || {
				let sel_files = self2.lst_files.columns().selected_texts(0);

				if sel_files.is_empty() {
					util::prompt::err(self2.wnd.hwnd(), "No files", None,
						"There are no selected files to be modified.")?;
					return Ok(());
				}

				let pop = WndModify::new(&self2.wnd, self2.tags_cache.clone(), Rc::new(sel_files));
				pop.show()?;

				{
					let tags_cache = self2.tags_cache.borrow();
					for idx in self2.lst_files.items().selected().iter() {
						let sel_file = self2.lst_files.items().text(*idx, 0);
						let tag = tags_cache.get(&sel_file).unwrap();
						self2.lst_files.items().set_text(*idx, 1, // write new padding
							&format!("{}", tag.original_padding()))?;
					}
				}

				self2.show_selected_tag_frames()?;
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_CLR_DIACR, {
			let self2 = self.clone();
			move || {
				let clock = util::Timer::start()?;
				let sel_idxs = self2.lst_files.items().selected();

				for idx in self2.lst_files.items().selected().iter() {
					let file = self2.lst_files.items().text(*idx, 0);
					let file_clean = util::clear_diacritics(&file); // generate clean name

					if file == file_clean { continue; } // if name didn't change, skip it

					{
						// Isolated scope because changing the item triggers
						// LVN_ITEMCHANGED, which will also borrow tags_cache.

						let mut tags_cache = self2.tags_cache.borrow_mut();
						let tag = tags_cache.remove(&file).unwrap();
						tags_cache.insert(file_clean.clone(), tag); // reinsert tag under clean name
					}

					self2.lst_files.items().set_text(*idx, 0, &file_clean)?; // change item
					w::MoveFile(&file, &file_clean)?; // rename file on disk
				}

				util::prompt::info(self2.wnd.hwnd(),
					"Operation successful", Some("Success"),
					&format!("Diacritics removed from {} file name(s) in {:.2} ms.",
						sel_idxs.len(), clock.now_ms()?))?;
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(id::MNU_FILE_ABOUT, {
			let self2 = self.clone();
			move || {
				let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
				let ri = w::ResourceInfo::read_from(&exe_name)?;
				let ver = ri.fixed_file_info().unwrap().dwFileVersion();
				let (lang, cp) = ri.langs_and_code_pages().unwrap()[0];

				util::prompt::info(self2.wnd.hwnd(),
					"About",
					Some(&format!("{} v{}.{}.{}",
						ri.product_name(lang, cp).unwrap(), ver[0], ver[1], ver[2])),
					&format!("Writen in Rust with WinSafe library.\n{}",
						ri.legal_copyright(lang, cp).unwrap()),
				)?;
				Ok(())
			}
		});
	}
}

