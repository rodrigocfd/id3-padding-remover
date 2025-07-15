use winsafe::{self as w, co, gui, prelude::*};

use super::{DlgEdit, DlgMain, LIST_COLS};
use crate::{ids, msgbox};

impl DlgMain {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			let lv = &self2.lst_files;
			self2.update_num_files_in_titlebar(lv.items().count())?;

			// Setup the files listview.
			lv.image_list(co::LVSIL::SMALL)?
				.add_icons_from_shell(&["mp3"])?;
			lv.set_extended_style(true, co::LVS_EX::FULLROWSELECT);
			lv.context_menu()
				.unwrap()
				.SetMenuDefaultItem(w::IdPos::Id(ids::MNU_FILE_EDIT))?;
			LIST_COLS.iter().try_for_each(|(title, cx, _)| {
				lv.cols().add(*title, gui::dpi_x(*cx))?; // add the columns
				w::SysResult::Ok(())
			})?;

			// Set files listview columns justification.
			let hcols = lv.header().unwrap().items();
			[1, 5, 8]
				.into_iter() // padding, track #, year
				.for_each(|i| {
					hcols.get(i).set_justify(gui::HeaderJustify::Right);
				});
			[2, 3]
				.into_iter() // art, RG
				.for_each(|i| {
					hcols.get(i).set_justify(gui::HeaderJustify::Center);
				});

			hcols.get(0).set_arrow(gui::HeaderArrow::Asc); // initially 1st col, ascending
			lv.cols().get(0).set_width_to_fill()?;
			self2.wnd.hwnd().RegisterDragDrop(&self2.drop_target)?;
			Ok(true)
		});

		let self2 = self.clone();
		self.wnd.on().wm_size(move |p| {
			if p.request != co::SIZE_R::MINIMIZED {
				self2.lst_files.cols().get(0).set_width_to_fill()?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_init_menu_popup(move |p| {
			if p.hmenu == self2.lst_files.context_menu().unwrap() {
				let has_sel = self2.lst_files.items().selected_count() >= 1;
				[
					ids::MNU_FILE_EDIT,
					ids::MNU_FILE_REMOVE,
					ids::MNU_FILE_RENAME_TAT,
					ids::MNU_FILE_RENAME_AT,
					ids::MNU_FILE_REWRITE,
					ids::MNU_FILE_WRITE_TRACK_NO,
					ids::MNU_FILE_REMRG,
					ids::MNU_FILE_REMRGART,
				]
				.into_iter()
				.try_for_each(|id| {
					p.hmenu
						.EnableMenuItem(w::IdPos::Id(id), has_sel)
						.map(|_| ())
				})?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_OPEN, move || {
				let fod = w::CoCreateInstance::<w::IFileOpenDialog>(
					&co::CLSID::FileOpenDialog,
					None::<&w::IUnknown>,
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
			.wm_command_acc_menu(ids::MNU_FILE_EDIT, move || {
				if self2.lst_files.items().selected_count() == 0 {
					return Ok(()); // Enter key will hit here even if there are no selected items
				}

				let cloned_sel_tags = self2
					.lst_files
					.items()
					.iter_selected()
					.map(|sel_item| {
						let rc_tag = sel_item.data()?;
						let cloned_tag = rc_tag.try_borrow()?.clone();
						Ok(cloned_tag)
					})
					.collect::<w::AnyResult<Vec<_>>>()?; // deep copy of selected tags

				// Show the modal window, which will take ownership of the cloned tags.
				// If user clicked OK, returns Some with the modified tags.
				if let Some(edited_tags) = DlgEdit::show(&self2.wnd, cloned_sel_tags)? {
					self2.lst_files.set_redraw(false);

					self2
						.lst_files
						.items()
						.iter_selected() // the items order should be the same
						.zip(edited_tags.into_iter())
						.try_for_each(|(sel_item, mut edited_tag)| -> w::AnyResult<()> {
							edited_tag.save_to_file(&sel_item.text(0))?;
							*sel_item.data()?.try_borrow_mut()? = edited_tag; // replace the tag currently stored in the item
							Self::render_tag(sel_item)?; // update the listview with the new values
							Ok(())
						})?;

					self2.lst_files.set_redraw(true);
				}

				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_REMOVE, move || {
				self2.lst_files.items().delete_selected()?;
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_REWRITE, move || {
				let count = self2.lst_files.items().selected_count();
				let ss = if count == 1 { "" } else { "s" };
				let content = format!("Rewrite the tag in {} file{}?", count, ss);

				if msgbox::ask(
					&self2.wnd,
					&format!("Rewrite file{}", ss),
					None,
					&content,
					"&Rewrite",
				)? {
					self2.lst_files.items().iter_selected().try_for_each(
						|sel_item| -> w::AnyResult<()> {
							{
								let rc_tag = sel_item.data()?; // retrieve tag saved in the listview item
								let mut tag = rc_tag.try_borrow_mut()?;
								tag.save_to_file(&sel_item.text(0))?; // save to MP3 file
							}
							Self::render_tag(sel_item)?; // padding will be set to zero, if any
							Ok(())
						},
					)?;
				}
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_WRITE_TRACK_NO, move || {
				self2
					.lst_files
					.items()
					.iter_selected()
					.enumerate()
					.try_for_each(|(idx, sel_item)| -> w::AnyResult<()> {
						{
							let rc_tag = sel_item.data()?;
							let mut tag = rc_tag.try_borrow_mut()?;
							tag.set_editable_string("TRCK", &(idx + 1).to_string())?;
							tag.save_to_file(&sel_item.text(0))?;
						}
						Self::render_tag(sel_item)?;
						Ok(())
					})?;

				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_RENAME_TAT, move || self2.rename(true));

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_RENAME_AT, move || self2.rename(false));

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_REMRG, move || self2.remove_rg_art(false));

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_REMRGART, move || self2.remove_rg_art(true));

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FILE_ABOUT, move || {
				let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
				let hversion = w::HVERSIONINFO::GetFileVersionInfo(&exe_name)?;
				let version_parts = hversion.version_info()?.dwFileVersion();

				let content = format!(
					"Version {}.{}.{}\n\
					Written in Rust with WinSafe library.\n\n\
					{}",
					version_parts[0],
					version_parts[1],
					version_parts[2],
					hversion.str_val(hversion.langs_and_cps()?[0], "LegalCopyright")?,
				);

				msgbox::info(&self2.wnd, "About", Some("ID3 Fit"), &content)?;
				Ok(())
			});

		let self2 = self.clone();
		self.lst_files.on().lvn_item_changed(move |_| {
			self2.update_num_files_in_titlebar(self2.lst_files.items().count())?;
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_key_down(move |p| {
			if p.wVKey == co::VK::DELETE {
				// On DEL key, remove selected files from the list.
				self2.lst_files.items().delete_selected()?;
			} else if p.wVKey == co::VK::RETURN {
				// On Enter key, edit the selected tags.
				self2
					.wnd
					.hwnd()
					.SendCommand(w::AccelMenuCtrl::Menu(ids::MNU_FILE_EDIT));
			}
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().nm_dbl_clk(move |_| {
			self2
				.wnd
				.hwnd()
				.SendCommand(w::AccelMenuCtrl::Menu(ids::MNU_FILE_EDIT));
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files.on().lvn_delete_item(move |_| {
			// Notification is sent before the list is updated.
			self2.update_num_files_in_titlebar(self2.lst_files.items().count() - 1)?;
			Ok(())
		});

		let self2 = self.clone();
		self.lst_files
			.header()
			.unwrap()
			.on()
			.hdn_item_click(move |p| {
				let new_col = p.iItem as u32; // index of column clicked by user
				let (cur_col, cur_is_asc) = self2.cur_sort.get(); // read current sort state
				let new_is_asc = new_col != cur_col || !cur_is_asc; // will sorting be ascending?
				self2.cur_sort.set((new_col, new_is_asc)); // save new sort state

				let cols = self2.lst_files.header().unwrap().items();
				cols.iter()?.for_each(|col| {
					col.set_arrow(gui::HeaderArrow::None); // remove arrow from all listview cols
				});
				let new_arrow =
					if new_is_asc { gui::HeaderArrow::Asc } else { gui::HeaderArrow::Desc };
				cols.get(new_col).set_arrow(new_arrow); // draw arrow in current col

				self2.sort_list()?;
				Ok(())
			});

		let self2 = self.clone();
		self.drop_target.Drop(move |d, _, _, _| {
			let mut fmt = w::FORMATETC::default();
			fmt.cfFormat = co::CF::HDROP;
			fmt.dwAspect = co::DVASPECT::CONTENT;
			fmt.tymed = co::TYMED::HGLOBAL;

			let medium = unsafe { d.GetData(&fmt)? };
			let hglobal = unsafe { medium.ptr_hglobal().unwrap() };
			let ptr_lock = hglobal.GlobalLock()?;
			let hdrop = unsafe { w::HDROP::from_ptr(ptr_lock.as_ptr() as _) };
			let dropped_paths = hdrop.DragQueryFile()?.collect::<w::SysResult<Vec<_>>>()?;

			self2.add_files_to_list(&dropped_paths)?;
			Ok(())
		});
	}
}
