use winsafe::{self as w, prelude::*, gui, msg};

use super::WndEdit;

impl WndEdit {
	pub(super) fn wm_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.on_init_dialog()
		});

		let self2 = self.clone();
		self.btn_uncheck.on().bn_clicked(move || {
			self2.inputs.try_borrow()?
				.iter()
				.for_each(|input| {
					input.chk.set_check_state_and_trigger(gui::CheckState::Unchecked);
				});
			Ok(())
		});

		let self2 = self.clone();
		self.btn_ok.on().bn_clicked(move || {
			self2.inputs.try_borrow()?
				.iter()
				.try_for_each(|input| {
					if input.chk.is_checked() { // field is checked?
						self2.sel_tags.iter()
							.try_for_each(|rc_tag| { // for each MP3 being edited
								let mut tag = rc_tag.try_borrow_mut()?;
								tag.set_frame_str(&input.name4, input.txt.text().trim())?;
								w::AnyResult::Ok(())
							})?;
					}
					w::AnyResult::Ok(())
				})?;
			self2.modal_return.set(true);
			unsafe { self2.wnd.hwnd().PostMessage(msg::wm::Close {}).ok(); }
			Ok(())
		});

		let self2 = self.clone();
		self.btn_cancel.on().bn_clicked(move || { // will also fire on Esc
			self2.modal_return.set(false);
			unsafe { self2.wnd.hwnd().PostMessage(msg::wm::Close {}).ok(); }
			Ok(())
		});

		self.inputs.borrow()
			.iter()
			.for_each(|input| {
				let fp2 = input.clone();
				input.chk.on().bn_clicked(move || { // event on each checkbox
					if fp2.chk.is_checked() {
						fp2.txt.hwnd().EnableWindow(true);
						fp2.txt.focus();
					} else {
						fp2.txt.hwnd().EnableWindow(false);
					}
					Ok(())
				});
			});
	}
}
