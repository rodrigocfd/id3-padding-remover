use winsafe::{self as w, co, gui, prelude::*};

use super::{DlgMain, LIST_COLS};
use crate::ids;

impl DlgMain {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.update_num_files(self2.lst_files.items().count())?;

			self2.lst_files.set_image_list(co::LVSIL::SMALL, {
				let il = w::HIMAGELIST::Create(w::SIZE::new(16, 16), co::ILC::COLOR32, 1, 1)?;
				il.add_icons_from_shell(&["mp3"])?;
				il
			});
			self2
				.lst_files
				.context_menu()
				.unwrap()
				.SetMenuDefaultItem(w::IdPos::Id(ids::MNU_MAIN_EDIT))?;
			self2
				.lst_files
				.set_extended_style(true, co::LVS_EX::FULLROWSELECT);
			LIST_COLS.iter().try_for_each(|(title, cx, _)| {
				self2.lst_files.cols().add(*title, gui::dpi_x(*cx))?;
				w::SysResult::Ok(())
			})?;

			[1, 5, 8]
				.iter() // padding, track #, year
				.for_each(|i| {
					self2
						.lst_files
						.header()
						.unwrap()
						.items()
						.get(*i)
						.set_justify(gui::HeaderJustify::Right);
				});
			[2, 3]
				.iter() // art, RG
				.for_each(|i| {
					self2
						.lst_files
						.header()
						.unwrap()
						.items()
						.get(*i)
						.set_justify(gui::HeaderJustify::Center);
				});

			self2.sort_list(0, true)?; // sort by path initially
			Ok(true)
		});

		let self2 = self.clone();
		self.wnd.on().wm_drop_files(move |p| {
			let dropped_files = p.hdrop.DragQueryFile()?.collect::<w::SysResult<Vec<_>>>()?;
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
		self.wnd
			.on()
			.wm_command_accel_menu(ids::MNU_MAIN_OPEN, move || {
				let fod = w::CoCreateInstance::<w::IFileOpenDialog>(
					&co::CLSID::FileOpenDialog,
					None,
					co::CLSCTX::INPROC_SERVER,
				)?;

				fod.SetOptions(
					fod.GetOptions()?
						| co::FOS::FORCEFILESYSTEM
						| co::FOS::FILEMUSTEXIST
						| co::FOS::ALLOWMULTISELECT,
				)?;

				fod.SetFileTypes(&[("MP3 audio files", "*.mp3"), ("All files", "*.*")])?;
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
		self.wnd
			.on()
			.wm_command_accel_menu(ids::MNU_MAIN_EDIT, move || {
				self2.edit_selected()?;
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_accel_menu(ids::MNU_MAIN_REMOVE, move || {
				self2.lst_files.items().delete_selected()?;
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_accel_menu(ids::MNU_MAIN_STRIP_RG, move || {
				self2.strip_replaygain_art(false)?;
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_accel_menu(ids::MNU_MAIN_STRIP_RG_ART, move || {
				self2.strip_replaygain_art(true)?;
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_accel_menu(ids::MNU_MAIN_ABOUT, move || {
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
					flags: co::TDF::ALLOW_DIALOG_CANCELLATION
						| co::TDF::POSITION_RELATIVE_TO_WINDOW,
					content: Some(&format!(
						"Version {}.{}.{}\n\
					Writen in Rust with WinSafe library.\n\n\
					{}",
						version_parts[0],
						version_parts[1],
						version_parts[2],
						hversion.str_val(lang0, cp0, "LegalCopyright")?,
					)),
					..Default::default()
				})?;
				Ok(())
			});

		let self2 = self.clone();
		self.wnd.on().wm_init_menu_popup(move |p| {
			if self2.lst_files.context_menu().unwrap() == p.hmenu {
				[
					ids::MNU_MAIN_EDIT,
					ids::MNU_MAIN_REMOVE,
					ids::MNU_MAIN_STRIP_RG,
					ids::MNU_MAIN_STRIP_RG_ART,
				]
				.into_iter()
				.try_for_each(|id| {
					p.hmenu
						.EnableMenuItem(
							w::IdPos::Id(id),
							self2.lst_files.items().selected_count() > 0, // at least 1 file selected?
						)
						.map(|_| ())
				})?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_item_changed(move |_| {
			self2.update_num_files(self2.lst_files.items().count())?;
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_key_down(move |p| {
			if p.wVKey == co::VK::DELETE {
				self2.lst_files.items().delete_selected()?; // on DEL key, remove selected files from the list
			} else if p.wVKey == co::VK::RETURN {
				self2.edit_selected()?; // on Enter key, edit the selected tags
			}
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().nm_dbl_clk(move |_| {
			self2.edit_selected()?;
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_delete_item(move |_| {
			// Notification is sent before the list is updated.
			self2.update_num_files(self2.lst_files.items().count() - 1)?;
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files
			.header()
			.unwrap()
			.on()
			.hdn_item_click(move |p| {
				self2.sort_list(p.iItem as u32, false)?;
				Ok(())
			});
	}
}
