use winsafe::{self as w, prelude::*, co};

use crate::{ids, wnd_edit::WndEdit};
use super::WndMain;

impl WndMain {
	pub(super) fn wm_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.init_dialog()
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
			if self2.lst_files.items().selected_count() == 0 {
				return Ok(()); // Enter key will hit here even if there are no selected items
			}

			let wnd_edit = WndEdit::new(
				&self2.wnd,
				self2.all_tags.clone(), // pointer to all tags in memory
				self2.lst_files.items() // currently selected MP3 paths
					.iter_selected()
					.map(|sel_item| sel_item.text(0))
					.collect(),
			);
			wnd_edit.show()?;

			self2.lst_files.set_redraw(false);
			{
				let all_tags = self2.all_tags.try_borrow()?;
				self2.lst_files.items() // update the listview values of all selected tags
					.iter_selected()
					.for_each(|sel_item| {
						let mp3_path = sel_item.text(0);
						all_tags.iter()
							.find(|path_and_tag| path_and_tag.mp3_path == mp3_path)
							.map(|path_and_tag| self2.write_tag_to_listview(sel_item, &path_and_tag.tag));
					});
			}
			self2.lst_files.set_redraw(true);
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_REMOVE, move || {
			self2.remove_selected_files()?;
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_accel_menu(ids::MNU_MAIN_ABOUT, move || {
			let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
			let hversion = w::HVERSIONINFO::GetFileVersionInfo(&exe_name)?;
			let version_parts = hversion.version_info()?.dwFileVersion();

			self2.wnd.hwnd().TaskDialog(
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
