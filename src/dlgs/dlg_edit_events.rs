use winsafe::{self as w, cmd, co, evp, evt, gui, prelude::*};

use super::{DlgEdit, ids};
use crate::msgbox;

impl DlgEdit {
	pub(super) fn events(&self) {
		evp!(self, wm_init_dialog, on_init_dialog);
		evp!(self, wm_init_menu_popup, on_init_menu_popup);
		cmd!(self, ids::MNU_FRAMES_MOVEUP, on_menu_frames_moveup);
		cmd!(self, ids::MNU_FRAMES_MOVEDOWN, on_menu_frames_movedown);
		cmd!(self, ids::MNU_FRAMES_DELETE, on_menu_frames_delete);
		evp!(self, self.lst_frames, lvn_key_down, on_lstframes_keydown);
		evt!(self, self.chk_pic, bn_clicked, on_chk_pick);
		evt!(self, self.btn_uncheck_all, bn_clicked, on_btn_unckeck_all);
		evt!(self, self.btn_check_filled, bn_clicked, on_btn_check_filled);
		cmd!(self, co::DLGID::OK, on_ok);
		cmd!(self, co::DLGID::CANCEL, on_esc);

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
	}

	fn on_init_dialog(&self, _: w::msg::WmInitDialog) -> w::AnyResult<bool> {
		// Add all the genres to the ComboBox.
		self.load_combo_genres()?;

		// For each checkbox + input, check if the frame has the same value across all tags.
		{
			let sel_tags = self.sel_tags.try_borrow()?;
			self.inputs
				.iter()
				.try_for_each(|input| input.set_text_if_equal_in_tags(&sel_tags))?;
		}

		// Setup the frames listview.
		let lv = &self.lst_frames;
		lv.cols().add("Frame", gui::dpi_x(56))?;
		lv.cols()
			.add("Value", gui::dpi_x(100))?
			.set_width_to_fill()?;
		lv.set_extended_style(true, co::LVS_EX::FULLROWSELECT | co::LVS_EX::GRIDLINES);

		self.render_frames_list()?;
		self.load_picture()?;
		self.chk_pic.hwnd().EnableWindow(false); // to be implemented later
		Ok(true)
	}

	fn on_init_menu_popup(&self, p: w::msg::WmInitMenuPopup) -> w::AnyResult<()> {
		let lv = &self.lst_frames;
		if p.hmenu == lv.context_menu().unwrap() {
			let has_sel = lv.items().selected_count() >= 1;
			let first_is_sel = lv.items().get(0).is_selected();
			let last_is_sel = lv.items().last().unwrap().is_selected();

			p.hmenu
				.EnableMenuItem(w::IdPos::Id(ids::MNU_FRAMES_MOVEUP), has_sel && !first_is_sel)?;
			p.hmenu
				.EnableMenuItem(w::IdPos::Id(ids::MNU_FRAMES_MOVEDOWN), has_sel && !last_is_sel)?;
			p.hmenu
				.EnableMenuItem(w::IdPos::Id(ids::MNU_FRAMES_DELETE), has_sel)?;
		}
		Ok(())
	}

	fn on_menu_frames_moveup(&self) -> w::AnyResult<()> {
		let items = self.lst_frames.items();
		let focus_idx = items.focused().map(|item| item.index());

		let new_sel_indexes = items // indexes of items that will appear as selected after the move
			.iter_selected()
			.map(|sel_item| {
				let idx = sel_item.index() as usize;
				self.sel_tags.try_borrow_mut()?[0] // assume we have only 1 MP3 loaded
					.frames_mut()
					.swap(idx, idx - 1); // swap frames in tag
				Ok(idx - 1)
			})
			.collect::<w::AnyResult<Vec<_>>>()?;

		self.render_frames_list()?;

		items.select_all(false)?;
		new_sel_indexes
			.iter()
			.try_for_each(|new_sel_idx| -> w::AnyResult<_> {
				items.get(*new_sel_idx as _).select(true)?; // re-select the moved items
				Ok(())
			})?;

		focus_idx.map(|idx| items.get(idx - 1).focus());
		Ok(())
	}

	fn on_menu_frames_movedown(&self) -> w::AnyResult<()> {
		let items = self.lst_frames.items();
		let focus_idx = items.focused().map(|item| item.index());

		let new_sel_indexes = items // indexes of items that will appear as selected after the move
			.iter_selected()
			.rev()
			.map(|sel_item| {
				let idx = sel_item.index() as usize;
				self.sel_tags.try_borrow_mut()?[0] // assume we have only 1 MP3 loaded
					.frames_mut()
					.swap(idx, idx + 1); // swap frames in tag
				Ok(idx + 1)
			})
			.collect::<w::AnyResult<Vec<_>>>()?;

		self.render_frames_list()?;

		items.select_all(false)?;
		new_sel_indexes
			.iter()
			.try_for_each(|new_sel_idx| -> w::AnyResult<_> {
				items.get(*new_sel_idx as _).select(true)?; // re-select the moved items
				Ok(())
			})?;

		focus_idx.map(|idx| items.get(idx + 1).focus());
		Ok(())
	}

	fn on_menu_frames_delete(&self) -> w::AnyResult<()> {
		if self.lst_frames.items().selected_count() == 0 {
			// Must check because listview's keydown will pop here even if
			// there are no selected items.
			return Ok(());
		}

		let msg = format!("Delete {} selected frame(s)?", self.lst_frames.items().selected_count());
		if msgbox::ask(&self.wnd, "Delete frame(s)", None, &msg, "&Delete")? {
			self.lst_frames.items().iter_selected().rev().try_for_each(
				|sel_item| -> w::AnyResult<_> {
					// Remove the frame by index directly from tag.
					self.sel_tags.try_borrow_mut()?[0] // assume we have only 1 MP3 loaded
						.frames_mut()
						.remove(sel_item.index() as _);
					Ok(())
				},
			)?;

			self.render_frames_list()?;
		}
		Ok(())
	}

	fn on_lstframes_keydown(&self, p: &w::NMLVKEYDOWN) -> w::AnyResult<()> {
		if p.wVKey == co::VK::DELETE {
			self.wnd
				.hwnd()
				.SendCommand(w::AccelMenuCtrl::Menu(ids::MNU_FRAMES_DELETE));
		}
		Ok(())
	}

	fn on_chk_pick(&self) -> w::AnyResult<()> {
		if self.chk_pic.is_checked() {
			self.wnd_pic.load_picture(&self.sel_tags.try_borrow()?)?; // ask the control to load the IPicture
		} else {
			self.wnd_pic.unload_picture()?;
		}
		Ok(())
	}

	fn on_btn_unckeck_all(&self) -> w::AnyResult<()> {
		self.inputs
			.iter()
			.try_for_each(|input| input.chk.set_check_and_trigger(false))?;
		self.chk_pic.set_check(false);
		Ok(())
	}

	fn on_btn_check_filled(&self) -> w::AnyResult<()> {
		self.inputs
			.iter()
			.try_for_each(|input| -> w::AnyResult<_> {
				let text = input.txt.hwnd().GetWindowText()?;
				if !text.is_empty() {
					input.chk.set_check(true);
					input.txt.hwnd().EnableWindow(true);
				}
				Ok(())
			})?;

		let has_pic = self.wnd_pic.pic.try_borrow()?.is_some();
		self.chk_pic.set_check(has_pic);
		Ok(())
	}

	fn on_ok(&self) -> w::AnyResult<()> {
		let mut sel_tags = self.sel_tags.try_borrow_mut()?;

		if sel_tags.len() == 1 {
			// If we're editing a single MP3 file, replace the frames with the
			// current ones in the listview.
			sel_tags[0].frames_mut().clear(); // delete all frames
			self.lst_frames
				.items()
				.iter()
				.try_for_each(|item| -> w::AnyResult<_> {
					let rc_frame = item.data();
					let cloned_frame = rc_frame.try_borrow()?.clone();
					sel_tags[0].frames_mut().push(cloned_frame); // add the frame from the listview
					Ok(())
				})?;
		}

		// Run through all textboxes and set/remove the text values on all tags.
		self.inputs
			.iter()
			.filter(|input| input.chk.is_checked())
			.try_for_each(|input| -> w::AnyResult<_> {
				let text = input.txt.hwnd().GetWindowText()?;
				sel_tags
					.iter_mut()
					.try_for_each(|tag| tag.set_editable_string(&input.name4, &text))?;
				Ok(())
			})?;

		self.user_clicked_ok.set(true);
		self.wnd.close();
		Ok(())
	}

	fn on_esc(&self) -> w::AnyResult<()> {
		self.user_clicked_ok.set(false);
		self.wnd.close();
		Ok(())
	}
}
