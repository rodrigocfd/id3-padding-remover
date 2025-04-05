use winsafe::{self as w, co, gui, msg, prelude::*};

use super::{DlgEdit, DlgMain, LIST_COLS};
use crate::{ids, msgbox};

impl DlgMain {
	pub(super) fn on_init_dialog(&self) -> w::AnyResult<bool> {
		self.update_num_files(self.lst_files.items().count())?;

		// Setup the files listview.
		self.lst_files.set_image_list(co::LVSIL::SMALL, {
			let il = w::HIMAGELIST::Create(w::SIZE::new(16, 16), co::ILC::COLOR32, 1, 1)?;
			il.add_icons_from_shell(&["mp3"])?;
			il
		});
		self.lst_files
			.set_extended_style(true, co::LVS_EX::FULLROWSELECT);
		self.lst_files
			.context_menu()
			.unwrap()
			.SetMenuDefaultItem(w::IdPos::Id(ids::MNU_FILE_EDIT))?;
		LIST_COLS
			.iter()
			.try_for_each(|(title, cx, _)| -> w::SysResult<()> {
				self.lst_files.cols().add(*title, gui::dpi_x(*cx))?;
				Ok(())
			})?;

		// Set files listview columns justification.
		[1, 5, 8]
			.iter() // padding, track #, year
			.for_each(|i| {
				self.lst_files
					.header()
					.unwrap()
					.items()
					.get(*i)
					.set_justify(gui::HeaderJustify::Right);
			});
		[2, 3]
			.iter() // art, RG
			.for_each(|i| {
				self.lst_files
					.header()
					.unwrap()
					.items()
					.get(*i)
					.set_justify(gui::HeaderJustify::Center);
			});

		self.wnd.hwnd().RegisterDragDrop(&self.drop_target)?;
		Ok(true)
	}

	pub(super) fn on_size(&self, p: msg::wm::Size) -> w::AnyResult<()> {
		if p.request != co::SIZE_R::MINIMIZED {
			self.lst_files.cols().get(0).set_width_to_fill()?;
		}
		Ok(())
	}

	pub(super) fn on_init_menu_popup(&self, p: msg::wm::InitMenuPopup) -> w::AnyResult<()> {
		if p.hmenu == self.lst_files.context_menu().unwrap() {
			let has_sel = self.lst_files.items().selected_count() >= 1;
			[
				ids::MNU_FILE_EDIT,
				ids::MNU_FILE_REMOVE,
				ids::MNU_FILE_RESAVE,
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
	}

	pub(super) fn on_menu_file_open(&self) -> w::AnyResult<()> {
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

		if fod.Show(self.wnd.hwnd())? {
			self.add_files_to_list(
				&fod.GetResults()?
					.iter()?
					.map(|shi| shi?.GetDisplayName(co::SIGDN::FILESYSPATH))
					.collect::<w::HrResult<Vec<_>>>()?,
			)?;
		}
		Ok(())
	}

	pub(super) fn on_menu_file_edit(&self) -> w::AnyResult<()> {
		if self.lst_files.items().selected_count() == 0 {
			return Ok(()); // Enter key will hit here even if there are no selected items
		}

		let cloned_sel_tags = self
			.lst_files
			.items()
			.iter_selected()
			.map(|sel_item| {
				let rc_tag = sel_item.data()?;
				let cloned_tag = rc_tag.try_borrow()?.clone();
				Ok(cloned_tag)
			})
			.collect::<w::AnyResult<Vec<_>>>()?; // deep copy of selected tags

		if let Some(edited_tags) = DlgEdit::show(&self.wnd, cloned_sel_tags)? {
			self.lst_files.set_redraw(false);

			self.lst_files
				.items()
				.iter_selected() // the items order should be the same
				.zip(edited_tags.into_iter())
				.try_for_each(|(sel_item, mut edited_tag)| -> w::AnyResult<()> {
					edited_tag.save_to_file(&sel_item.text(0))?;
					*sel_item.data()?.try_borrow_mut()? = edited_tag; // replace the tag currently stored in the item
					Self::render_tag(sel_item)?; // update the listview with the new values
					Ok(())
				})?;

			self.lst_files.set_redraw(true);
		}

		Ok(())
	}

	pub(super) fn on_menu_file_remove(&self) -> w::AnyResult<()> {
		self.lst_files.items().delete_selected()?;
		Ok(())
	}

	pub(super) fn on_menu_file_resave(&self) -> w::AnyResult<()> {
		let count = self.lst_files.items().selected_count();
		let ss = if count == 1 { "" } else { "s" };
		let content = format!("Rewrite the tag in {} file{}?", count, ss);

		if msgbox::ask(self.wnd.hwnd(), &format!("Rewrite file{}", ss), None, &content, "&Rewrite")?
		{
			self.lst_files.items().iter_selected().try_for_each(
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
	}

	pub(super) fn on_menu_file_del_rg_art(&self, del_art: bool) -> w::AnyResult<()> {
		let sel_count = self.lst_files.items().selected_count();
		let window_title = if del_art { "Remove ReplayGain and art" } else { "Remove ReplayGain" };
		let content = format!(
			"Remove ReplayGain {} frames of {} tag{}?",
			if del_art { "and art" } else { "" },
			sel_count,
			if sel_count == 1 { "" } else { "s" },
		);

		if msgbox::ask(self.wnd.hwnd(), window_title, None, &content, "&Remove")? {
			self.lst_files.items().iter_selected().try_for_each(
				|sel_item| -> w::AnyResult<()> {
					{
						let rc_tag = sel_item.data()?; // retrieve tag saved in the listview item
						let mut tag = rc_tag.try_borrow_mut()?;
						tag.frames_mut().retain(|frame| !frame.is_replay_gain());
						if del_art {
							tag.frames_mut().retain(|frame| frame.name4() != "APIC");
						}
						tag.save_to_file(&sel_item.text(0))?; // save to MP3 file
					}
					Self::render_tag(sel_item)?;
					Ok(())
				},
			)?;
		}
		Ok(())
	}

	pub(super) fn on_menu_file_about(&self) -> w::AnyResult<()> {
		let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
		let hversion = w::HVERSIONINFO::GetFileVersionInfo(&exe_name)?;
		let version_parts = hversion.version_info()?.dwFileVersion();

		let content = format!(
			"Version {}.{}.{}\n\
					Writen in Rust with WinSafe library.\n\n\
					{}",
			version_parts[0],
			version_parts[1],
			version_parts[2],
			hversion.str_val(hversion.langs_and_cps()?[0], "LegalCopyright")?,
		);

		msgbox::info(self.wnd.hwnd(), "About", Some("ID3 Fit"), &content)?;
		Ok(())
	}

	pub(super) fn on_lst_files_item_changed(&self) -> w::AnyResult<()> {
		self.update_num_files(self.lst_files.items().count())?;
		Ok(())
	}

	pub(super) fn on_lst_files_key_down(&self, p: &w::NMLVKEYDOWN) -> w::AnyResult<()> {
		if p.wVKey == co::VK::DELETE {
			self.lst_files.items().delete_selected()?; // on DEL key, remove selected files from the list
		} else if p.wVKey == co::VK::RETURN {
			self.on_menu_file_edit()?; // on Enter key, edit the selected tags
		}
		Ok(())
	}

	pub(super) fn on_lst_files_delete_item(&self) -> w::AnyResult<()> {
		// Notification is sent before the list is updated.
		self.update_num_files(self.lst_files.items().count() - 1)?;
		Ok(())
	}

	pub(super) fn on_header_item_click(&self, p: &w::NMHEADER) -> w::AnyResult<()> {
		let new_col = p.iItem as u32;
		let (cur_col, cur_is_asc) = self.cur_sort.get(); // read current sort state
		let is_asc = new_col != cur_col || !cur_is_asc; // will sorting be ordinary, ascending?
		self.cur_sort.set((new_col, is_asc)); // save new sort state
		self.sort_list()?;
		Ok(())
	}

	pub(super) fn on_drop_target_drag_enter(&self, fx: &mut co::DROPEFFECT) -> w::AnyResult<()> {
		*fx &= co::DROPEFFECT::COPY;
		Ok(())
	}

	pub(super) fn on_drop_target_drag_over(&self, fx: &mut co::DROPEFFECT) -> w::AnyResult<()> {
		*fx &= co::DROPEFFECT::COPY;
		Ok(())
	}

	pub(super) fn on_drop_target_drop(
		&self,
		d: &w::IDataObject,
		fx: &mut co::DROPEFFECT,
	) -> w::AnyResult<()> {
		let mut fmt = w::FORMATETC::default();
		fmt.cfFormat = co::CF::HDROP;
		fmt.dwAspect = co::DVASPECT::CONTENT;
		fmt.tymed = co::TYMED::HGLOBAL;

		let medium = unsafe { d.GetData(&fmt)? };
		let hglobal = unsafe { medium.ptr_hglobal().unwrap() };
		let ptr_lock = hglobal.GlobalLock()?;
		let hdrop = unsafe { w::HDROP::from_ptr(ptr_lock.as_ptr() as _) };
		let dropped_paths = hdrop.DragQueryFile()?.collect::<w::SysResult<Vec<_>>>()?;

		self.add_files_to_list(&dropped_paths)?;

		*fx &= co::DROPEFFECT::COPY;
		Ok(())
	}
}
