use winsafe::{self as w, prelude::*, msg};

use super::WndEdit;

impl WndEdit {
	pub(super) fn wm_events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_init_dialog(move |_| {
			self2.init_dialog()
		});

		let self2 = self.clone();
		self.btn_ok.on().bn_clicked(move || {
			let mut all_tags = self2.all_tags.try_borrow_mut()?;
			let mut edited_tags = all_tags.iter_mut()
				.filter(|path_and_tag| self2.selected_paths.contains(&path_and_tag.mp3_path))
				.map(|path_and_tag| &mut path_and_tag.tag)
				.collect::<Vec<_>>();

			self2.field_packs.try_borrow()?
				.iter()
				.try_for_each(|field_pack| {
					if field_pack.chk.is_checked() { // field is checked?
						edited_tags.iter_mut() // for each MP3 being edited
							.try_for_each(|tag| {
								tag.set_known_field( // save the text to tag in Vec
									field_pack.field,
									field_pack.txt.text().trim(),
								)?;
								w::AnyResult::Ok(())
							})?;
					}
					w::AnyResult::Ok(())
				})?;
			self2.wnd.hwnd().PostMessage(msg::wm::Close {})?;
			Ok(())
		});

		let self2 = self.clone();
		self.btn_cancel.on().bn_clicked(move || { // will also fire on Esc
			self2.wnd.hwnd().PostMessage(msg::wm::Close {})?;
			Ok(())
		});

		self.field_packs.borrow()
			.iter()
			.for_each(|field_pack| {
				let fp2 = field_pack.clone();
				field_pack.chk.on().bn_clicked(move || {
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
