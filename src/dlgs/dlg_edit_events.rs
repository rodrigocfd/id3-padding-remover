use winsafe::{self as w, co, gui, prelude::*};

use super::{DlgEdit, Input};

impl DlgEdit {
	pub(super) fn on_init_dialog(&self) -> w::AnyResult<bool> {
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
		self.lst_frames.cols().add("Frame", gui::dpi_x(56))?;
		self.lst_frames
			.cols()
			.add("Value", gui::dpi_x(100))?
			.set_width_to_fill()?;
		self.lst_frames
			.set_extended_style(true, co::LVS_EX::FULLROWSELECT | co::LVS_EX::GRIDLINES);

		self.render_frames_list()?;
		self.load_picture()?;
		self.chk_pic.hwnd().EnableWindow(false); // to be implemented later
		Ok(true)
	}

	pub(super) fn on_chk_click(&self, input: &Input) -> w::AnyResult<()> {
		if input.chk.is_checked() {
			input.txt.hwnd().EnableWindow(true);
			input.txt.focus()?; // when checked, enable the textbox and focus it
		} else {
			input.txt.hwnd().EnableWindow(false);
		}
		Ok(())
	}

	pub(super) fn on_chk_pic_click(&self) -> w::AnyResult<()> {
		if self.chk_pic.is_checked() {
			self.wnd_pic.load_picture(&self.sel_tags.try_borrow()?)?; // ask the control to load the IPicture
		} else {
			self.wnd_pic.unload_picture()?;
		}
		Ok(())
	}

	pub(super) fn on_uncheck_all(&self) -> w::AnyResult<()> {
		self.inputs
			.iter()
			.try_for_each(|input| input.chk.set_check_and_trigger(false))?;
		self.chk_pic.set_check(false);
		Ok(())
	}

	pub(super) fn on_check_filled(&self) -> w::AnyResult<()> {
		self.inputs
			.iter()
			.try_for_each(|input| -> w::AnyResult<()> {
				let text = input.txt.hwnd().GetWindowText()?;
				if !text.is_empty() {
					input.chk.set_check(true);
					input.txt.hwnd().EnableWindow(true);
				}
				Ok(())
			})
	}

	pub(super) fn on_ok(&self) -> w::AnyResult<()> {
		self.inputs
			.iter()
			.filter(|input| input.chk.is_checked())
			.try_for_each(|input| -> w::AnyResult<()> {
				let text = input.txt.hwnd().GetWindowText()?;
				let mut sel_tags = self.sel_tags.try_borrow_mut()?;
				sel_tags
					.iter_mut()
					.try_for_each(|tag| tag.set_editable_string(&input.name4, &text))?;
				Ok(())
			})?;

		self.user_clicked_ok.set(true);
		self.wnd.close();
		Ok(())
	}

	pub(super) fn on_cancel(&self) -> w::AnyResult<()> {
		self.user_clicked_ok.set(false);
		self.wnd.close();
		Ok(())
	}
}
