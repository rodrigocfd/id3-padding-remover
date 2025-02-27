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
		self.wnd.on().wm_drop_files(move |p| {
			let dropped_files = p.hdrop.DragQueryFile()?
				.collect::<w::SysResult<Vec<_>>>()?;
			let mut valid_files = Vec::<String>::with_capacity(dropped_files.len());

			for file in dropped_files.iter() {
				if w::path::is_directory(file) {
					for sub_file in w::path::dir_list(file, Some("*.mp3")) {
						valid_files.push(sub_file?); // search only 1 level below
					}
				} else if w::path::has_extension(file, &[".mp3"]) {
					valid_files.push(file.clone());
				}
			}

			self2.add_files_to_list(&valid_files)?;
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_OPEN, move || {
			let fod = w::CoCreateInstance::<w::IFileOpenDialog>(
				&co::CLSID::FileOpenDialog, None, co::CLSCTX::INPROC_SERVER)?;

			fod.SetOptions(
				fod.GetOptions()?
					| co::FOS::FORCEFILESYSTEM
					| co::FOS::FILEMUSTEXIST
					| co::FOS::ALLOWMULTISELECT,
			)?;

			fod.SetFileTypes(&[
				("MP3 audio files", "*.mp3"),
				("All files", "*.*"),
			])?;
			fod.SetFileTypeIndex(1)?;

			if fod.Show(self2.wnd.hwnd())? {
				self2.add_files_to_list(
					&fod.GetResults()?
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
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_STRIP_RG, move || {
			self2.strip_replaygain_art(false)?;
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_STRIP_RG_ART, move || {
			self2.strip_replaygain_art(true)?;
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_ABOUT, move || {
			let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
			let hversion = w::HVERSIONINFO::GetFileVersionInfo(&exe_name)?;
			let (lang0, cp0) = hversion.langs_and_cps()?[0];
			let version_parts = hversion.version_info()?.dwFileVersion();

			w::TaskDialogIndirect(&w::TASKDIALOGCONFIG {
				hwnd_parent: Some(self2.wnd.hwnd()),
				window_title: Some("About"),
				main_instruction: Some("ID3 Fit"),
				main_icon: w::IconIdTd::Td(co::TD_ICON::INFORMATION),
				common_buttons: co::TDCBF::OK,
				flags: co::TDF::ALLOW_DIALOG_CANCELLATION | co::TDF::POSITION_RELATIVE_TO_WINDOW,
				content: Some(&format!(
					"Version {}.{}.{}\n\
					Writen in Rust with WinSafe library.\n\n\
					{}",
					version_parts[0], version_parts[1], version_parts[2],
					hversion.str_val(lang0, cp0, "LegalCopyright")?,
				)),
				..Default::default()
			})?;
			Ok(())
		});
	}
}
