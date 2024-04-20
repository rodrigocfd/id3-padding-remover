use winsafe::{self as w, prelude::*, gui};

use super::WndEdit;

impl WndEdit {
	pub(super) fn wm_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.on_init_dialog()
		});

		let self2 = self.clone();
		self.btn_uncheck.on().bn_clicked(move || {
			self2.field_packs.try_borrow()?
				.iter()
				.for_each(|field_pack| {
					field_pack.chk.set_check_state_and_trigger(gui::CheckState::Unchecked);
				});
			Ok(())
		});

		let self2 = self.clone();
		self.btn_ok.on().bn_clicked(move || {
			self2.field_packs.try_borrow()?
				.iter()
				.try_for_each(|field_pack| {
					if field_pack.chk.is_checked() { // field is checked?
						self2.sel_tags.iter()
							.try_for_each(|rc_tag| { // for each MP3 being edited
								let mut tag = rc_tag.try_borrow_mut()?;
								tag.set_frame_str(&field_pack.name4, field_pack.txt.text().trim())?;
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
		self.btn_cancel.on().bn_clicked(move || { // will also fire on Esc
			self2.modal_return.set(false);
			self2.wnd.close();
			Ok(())
		});

		self.field_packs.borrow()
			.iter()
			.for_each(|field_pack| {
				let fp2 = field_pack.clone();
				field_pack.chk.on().bn_clicked(move || { // event on each checkbox
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
