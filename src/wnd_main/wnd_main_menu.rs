use winsafe::{prelude::*, self as w, co, shell};

use crate::util;
use super::{ids, WndMain};

impl WndMain {
	pub(super) fn _menu_events(&self) {
		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_OPEN, {
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
					self2._add_files(
						&fileo.GetResults()?.iter()
							.map(|shi|
								shi.and_then(|shi|
									shi.GetDisplayName(shell::co::SIGDN::FILESYSPATH)
								),
							)
							.collect::<w::WinResult<Vec<_>>>()?,
					)?;
				}
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_DELETE, {
			let lst_files = self.lst_mp3s.clone();
			move || {
				lst_files.items().delete_selected()?;
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_REM_PAD, {
			let self2 = self.clone();
			move || {
				let clock = util::Timer::start()?;
				self2._remove_frames_from_sel_files_and_save(false, false)?;

				util::prompt::info(self2.wnd.hwnd(),
					"Operation successful", Some("Success"),
					&format!("Padding removed from {} file(s) in {:.2} ms.",
						self2.lst_mp3s.items().selected_count(), clock.now_ms()?))?;

				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_REM_RG, {
			let self2 = self.clone();
			move || {
				let clock = util::Timer::start()?;
				self2._remove_frames_from_sel_files_and_save(true, false)?;

				util::prompt::info(self2.wnd.hwnd(),
					"Operation successful", Some("Success"),
					&format!("ReplayGain removed from {} file(s) in {:.2} ms.",
						self2.lst_mp3s.items().selected_count(), clock.now_ms()?))?;

				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_REM_RG_PIC, {
			let self2 = self.clone();
			move || {
				let clock = util::Timer::start()?;
				self2._remove_frames_from_sel_files_and_save(true, true)?;

				util::prompt::info(self2.wnd.hwnd(),
					"Operation successful", Some("Success"),
					&format!("ReplayGain and album art removed from {} file(s) in {:.2} ms.",
						self2.lst_mp3s.items().selected_count(), clock.now_ms()?))?;

				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_RENAME, {
			let self2 = self.clone();
			move || {
				if let Err(err) = self2._rename_files(false) {
					util::prompt::err(self2.wnd.hwnd(), "Error", None, &err.to_string())?;
				}
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_RENAME_PREFIX, {
			let self2 = self.clone();
			move || {
				if let Err(err) = self2._rename_files(true) {
					util::prompt::err(self2.wnd.hwnd(), "Error", None, &err.to_string())?;
				}
				Ok(())
			}
		});

		self.wnd.on().wm_command_accel_menu(ids::MNU_MP3S_ABOUT, {
			let self2 = self.clone();
			move || {
				// let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
				// let res_info = w::ResourceInfo::read_from(&exe_name)?;
				// let ver = res_info.version_info().unwrap().dwFileVersion();
				// let block = res_info.blocks().next().unwrap(); // first block

				// util::prompt::info(self2.wnd.hwnd(),
				// 	"About",
				// 	Some(&format!("{} v{}.{}.{}",
				// 		block.product_name().unwrap(),
				// 		ver[0], ver[1], ver[2])),
				// 	&format!("Writen in Rust with WinSafe library.\n{}",
				// 		block.legal_copyright().unwrap()),
				// )?;

				Ok(())
			}
		});
	}
}

