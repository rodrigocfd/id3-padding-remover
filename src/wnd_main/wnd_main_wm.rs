use winsafe::{self as w, prelude::*, co};

use crate::ids;
use super::WndMain;

impl WndMain {
	pub(super) fn wm_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.on_init_dialog()
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_OPEN, move || {
			let fileo = w::CoCreateInstance::<w::IFileOpenDialog>(
				&co::CLSID::FileOpenDialog, None, co::CLSCTX::INPROC_SERVER)?;

			fileo.SetOptions(
				fileo.GetOptions()?
					| co::FOS::FORCEFILESYSTEM
					| co::FOS::FILEMUSTEXIST
					| co::FOS::ALLOWMULTISELECT,
			)?;

			fileo.SetFileTypes(&[
				("MP3 audio files", "*.mp3"),
			])?;
			fileo.SetFileTypeIndex(1)?;

			if fileo.Show(self2.wnd.hwnd())? {
				self2.add_files(
					&fileo.GetResults()?
						.iter()?
						.map(|shi| shi?.GetDisplayName(co::SIGDN::FILESYSPATH))
						.collect::<w::HrResult<Vec<_>>>()?,
				)?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_EDIT, move || {
			self2.edit_selected()?;
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_REMOVE, move || {
			self2.lst_files.items().delete_selected();
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_ABOUT, move || {
			let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
			let hversion = w::HVERSIONINFO::GetFileVersionInfo(&exe_name)?;
			let version_parts = hversion.version_info()?.dwFileVersion();

			self2.wnd.hwnd().TaskDialog(
				Some("About"),
				Some("ID3 Fit"),
				Some(&format!(
					"Version {}.{}.{}\n\
					Writen in Rust with WinSafe library.\n\n\
					Rodrigo César de Freitas Dias © 2024",
					version_parts[0], version_parts[1], version_parts[2],
				)),
				co::TDCBF::OK,
				w::IconRes::Info,
			)?;
			Ok(())
		});
	}
}
