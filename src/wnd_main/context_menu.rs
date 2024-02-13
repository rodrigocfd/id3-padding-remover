use winsafe::{self as w, prelude::*, co, gui, msg};

use crate::ids;
use super::WndMain;

impl WndMain {
	pub(super) fn context_menu(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_OPEN, move || {

			Ok(())
		});

		let wnd = self.wnd.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_ABOUT, move || {
			let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
			let hversion = w::HVERSIONINFO::GetFileVersionInfo(&exe_name)?;
			let version_info = hversion.version_info()?;
			let version_parts = version_info.dwFileVersion();

			wnd.hwnd().TaskDialog(
				None,
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
