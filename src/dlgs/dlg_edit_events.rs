use winsafe::{self as w, co, gui, prelude::*};

use super::DlgEdit;
use crate::{ids, msgbox};

impl DlgEdit {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			// Add all the genres to the ComboBox.
			self2.load_combo_genres()?;

			// For each checkbox + input, check if the frame has the same value across all tags.
			{
				let sel_tags = self2.sel_tags.try_borrow()?;
				self2
					.inputs
					.iter()
					.try_for_each(|input| input.set_text_if_equal_in_tags(&sel_tags))?;
			}

			// Setup the frames listview.
			let lv = &self2.lst_frames;
			lv.cols().add("Frame", gui::dpi_x(56))?;
			lv.cols()
				.add("Value", gui::dpi_x(100))?
				.set_width_to_fill()?;
			lv.set_extended_style(true, co::LVS_EX::FULLROWSELECT | co::LVS_EX::GRIDLINES);

			self2.render_frames_list()?;
			self2.load_picture()?;
			self2.chk_pic.hwnd().EnableWindow(false); // to be implemented later
			Ok(true)
		});

		let self2 = self.clone();
		self.wnd.on().wm_init_menu_popup(move |p| {
			let lv = &self2.lst_frames;
			if p.hmenu == lv.context_menu().unwrap() {
				let has_sel = lv.items().selected_count() >= 1;
				let first_is_sel = lv.items().get(0).is_selected();
				let last_is_sel = lv.items().last().unwrap().is_selected();

				p.hmenu.EnableMenuItem(
					w::IdPos::Id(ids::MNU_FRAMES_MOVEUP),
					has_sel && !first_is_sel,
				)?;
				p.hmenu.EnableMenuItem(
					w::IdPos::Id(ids::MNU_FRAMES_MOVEDOWN),
					has_sel && !last_is_sel,
				)?;
				p.hmenu
					.EnableMenuItem(w::IdPos::Id(ids::MNU_FRAMES_DELETE), has_sel)?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FRAMES_MOVEUP, move || {
				let items = self2.lst_frames.items();
				let focus_idx = items.focused().map(|item| item.index());

				let new_sel_indexes = items // indexes of items that will appear as selected after the move
					.iter_selected()
					.map(|sel_item| {
						let idx = sel_item.index() as usize;
						self2.sel_tags.try_borrow_mut()?[0] // assume we have only 1 MP3 loaded
							.frames_mut()
							.swap(idx, idx - 1); // swap frames in tag
						Ok(idx - 1)
					})
					.collect::<w::AnyResult<Vec<_>>>()?;

				self2.render_frames_list()?;

				items.select_all(false)?;
				new_sel_indexes
					.iter()
					.try_for_each(|new_sel_idx| -> w::AnyResult<_> {
						items.get(*new_sel_idx as _).select(true)?; // re-select the moved items
						Ok(())
					})?;

				focus_idx.map(|idx| items.get(idx - 1).focus());
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FRAMES_MOVEDOWN, move || {
				let items = self2.lst_frames.items();
				let focus_idx = items.focused().map(|item| item.index());

				let new_sel_indexes = items // indexes of items that will appear as selected after the move
					.iter_selected()
					.rev()
					.map(|sel_item| {
						let idx = sel_item.index() as usize;
						self2.sel_tags.try_borrow_mut()?[0] // assume we have only 1 MP3 loaded
							.frames_mut()
							.swap(idx, idx + 1); // swap frames in tag
						Ok(idx + 1)
					})
					.collect::<w::AnyResult<Vec<_>>>()?;

				self2.render_frames_list()?;

				items.select_all(false)?;
				new_sel_indexes
					.iter()
					.try_for_each(|new_sel_idx| -> w::AnyResult<_> {
						items.get(*new_sel_idx as _).select(true)?; // re-select the moved items
						Ok(())
					})?;

				focus_idx.map(|idx| items.get(idx + 1).focus());
				Ok(())
			});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(ids::MNU_FRAMES_DELETE, move || {
				if self2.lst_frames.items().selected_count() == 0 {
					// Must check because listview's keydown will pop here even if
					// there are no selected items.
					return Ok(());
				}

				let msg = format!(
					"Delete {} selected frame(s)?",
					self2.lst_frames.items().selected_count()
				);
				if msgbox::ask(&self2.wnd, "Delete frame(s)", None, &msg, "&Delete")? {
					self2
						.lst_frames
						.items()
						.iter_selected()
						.rev()
						.try_for_each(|sel_item| -> w::AnyResult<_> {
							// Remove the frame by index directly from tag.
							self2.sel_tags.try_borrow_mut()?[0] // assume we have only 1 MP3 loaded
								.frames_mut()
								.remove(sel_item.index() as _);
							Ok(())
						})?;

					self2.render_frames_list()?;
				}
				Ok(())
			});

		self.inputs.iter().for_each(|input| {
			let input2 = input.clone();
			input.chk.on().bn_clicked(move || {
				if input2.chk.is_checked() {
					input2.txt.hwnd().EnableWindow(true);
					input2.txt.focus()?; // when checked, enable the textbox and focus it
				} else {
					input2.txt.hwnd().EnableWindow(false);
				}
				Ok(())
			});
		});

		let self2 = self.clone();
		self.lst_frames.on().lvn_key_down(move |p| {
			if p.wVKey == co::VK::DELETE {
				self2
					.wnd
					.hwnd()
					.SendCommand(w::AccelMenuCtrl::Menu(ids::MNU_FRAMES_DELETE));
			}
			Ok(())
		});

		let self2 = self.clone();
		self.chk_pic.on().bn_clicked(move || {
			if self2.chk_pic.is_checked() {
				self2.wnd_pic.load_picture(&self2.sel_tags.try_borrow()?)?; // ask the control to load the IPicture
			} else {
				self2.wnd_pic.unload_picture()?;
			}
			Ok(())
		});

		let self2 = self.clone();
		self.btn_uncheck_all.on().bn_clicked(move || {
			self2
				.inputs
				.iter()
				.try_for_each(|input| input.chk.set_check_and_trigger(false))?;
			self2.chk_pic.set_check(false);
			Ok(())
		});

		let self2 = self.clone();
		self.btn_check_filled.on().bn_clicked(move || {
			self2
				.inputs
				.iter()
				.try_for_each(|input| -> w::AnyResult<_> {
					let text = input.txt.hwnd().GetWindowText()?;
					if !text.is_empty() {
						input.chk.set_check(true);
						input.txt.hwnd().EnableWindow(true);
					}
					Ok(())
				})?;

			let has_pic = self2.wnd_pic.pic.try_borrow()?.is_some();
			self2.chk_pic.set_check(has_pic);
			Ok(())
		});

		let self2 = self.clone();
		self.wnd.on().wm_command_acc_menu(co::DLGID::OK, move || {
			let mut sel_tags = self2.sel_tags.try_borrow_mut()?;

			if sel_tags.len() == 1 {
				// If we're editing a single MP3 file, replace the frames with the
				// current ones in the listview.
				sel_tags[0].frames_mut().clear(); // delete all frames
				self2
					.lst_frames
					.items()
					.iter()
					.try_for_each(|item| -> w::AnyResult<_> {
						let rc_frame = item.data()?;
						let cloned_frame = rc_frame.try_borrow()?.clone();
						sel_tags[0].frames_mut().push(cloned_frame); // add the frame from the listview
						Ok(())
					})?;
			}

			// Run through all textboxes and set/remove the text values on all tags.
			self2
				.inputs
				.iter()
				.filter(|input| input.chk.is_checked())
				.try_for_each(|input| -> w::AnyResult<_> {
					let text = input.txt.hwnd().GetWindowText()?;
					sel_tags
						.iter_mut()
						.try_for_each(|tag| tag.set_editable_string(&input.name4, &text))?;
					Ok(())
				})?;

			self2.user_clicked_ok.set(true);
			self2.wnd.close();
			Ok(())
		});

		let self2 = self.clone();
		self.wnd
			.on()
			.wm_command_acc_menu(co::DLGID::CANCEL, move || {
				self2.user_clicked_ok.set(false);
				self2.wnd.close();
				Ok(())
			});
	}
}
