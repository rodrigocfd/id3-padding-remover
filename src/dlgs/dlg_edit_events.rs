use winsafe::{self as w, prelude::*};

use super::DlgEdit;

impl DlgEdit {
	pub(super) fn wm_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.wnd.hwnd().SetWindowText(&format!(
				"Edit {} file{}",
				self2.sel_tags.len(),
				if self2.sel_tags.len() == 1 { "" } else { "s" },
			))?;
			self2.fill_chks_and_txts()?;
			self2.fill_listview_fields()?;
			Ok(true)
		});

		let self2 = self.clone();
		self.btn_uncheck.on().bn_clicked(move || {
			self2.inputs.try_borrow()?.iter().try_for_each(|input| {
				input.chk.set_check_and_trigger(false)?;
				w::SysResult::Ok(())
			})?;
			Ok(())
		});

		let self2 = self.clone();
		self.btn_ok.on().bn_clicked(move || {
			self2.inputs.try_borrow()?.iter().try_for_each(|input| {
				if input.chk.is_checked() {
					self2.sel_tags.iter().try_for_each(|rc_tag| {
						// For each MP3 being edited.
						let mut tag = rc_tag.try_borrow_mut()?;
						let text = input.txt.hwnd().GetWindowText()?;
						tag.set_frame_str(&input.name4, text.trim())?;
						w::AnyResult::Ok(())
					})?;
				}
				w::AnyResult::Ok(())
			})?;
			self2.modal_return.set(true);
			self2.wnd.close();
			Ok(())
		});

		let self2 = self.clone();
		self.btn_cancel.on().bn_clicked(move || {
			// Will also fire on Esc.
			self2.modal_return.set(false);
			self2.wnd.close();
			Ok(())
		});

		self.inputs.borrow().iter().for_each(|input| {
			let fp2 = input.clone();
			input.chk.on().bn_clicked(move || {
				// Event on each checkbox.
				if fp2.chk.is_checked() {
					fp2.txt.hwnd().EnableWindow(true);
					fp2.txt.focus()?;
				} else {
					fp2.txt.hwnd().EnableWindow(false);
				}
				Ok(())
			});
		});
	}
}
