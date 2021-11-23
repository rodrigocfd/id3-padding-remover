use winsafe::{prelude::*, self as w, gui, path};

use crate::id3v2;
use super::WndFields;

impl WndFields {
	pub(super) fn _events(&self) {
		self.wnd.on().wm_init_dialog({
			let cmb_genre = self.fields.iter()
				.find(|f| f.name == id3v2::FieldName::Genre).unwrap()
				.txt.as_any()
				.downcast_ref::<gui::ComboBox>().unwrap()
				.clone();

			move |_| {
				let genres_text = { // read genres from TXT
					let path = format!("{}\\id3-fit-genres.txt", path::exe_path()?);
					let fin = w::FileMapped::open(&path, w::FileAccess::ExistingReadOnly)?;
					w::WString::parse_str(fin.as_slice())?.to_string()
				};

				cmb_genre.items().add(&genres_text.lines().collect::<Vec<_>>())?;
				Ok(true)
			}
		});

		for field in self.fields.iter() {
			field.chk.on().bn_clicked({ // add event on each checkbox
				let (self2, field) = (self.clone(), field.clone());
				move || {
					field.chk.focus()?;
					field.txt.hwnd().EnableWindow(field.chk.is_checked());
					if field.chk.is_checked() {
						field.txt.focus()?;
					}
					self2._enable_buttons_if_at_least_one_checked();
					Ok(())
				}
			});
		}

		self.btn_clear_checks.on().bn_clicked({
			let self2 = self.clone();
			move || {
				for field in self2.fields.iter() {
					field.chk.set_check_state(gui::CheckState::Unchecked);
					field.txt.hwnd().EnableWindow(false);
				}
				self2._enable_buttons_if_at_least_one_checked();
				Ok(())
			}
		});

		self.btn_save.on().bn_clicked({
			let self2 = self.clone();
			move || {
				for field in self2.fields.iter() {
					if !field.chk.is_checked() { continue; } // skip unchecked textboxes

					let sel_mp3s = self2.sel_mp3s.try_borrow_mut()?;
					let new_text = field.txt.text()?.trim().to_owned(); // text typed by the user

					for (_, sel_tag) in self2.tags_cache.lock().unwrap()
						.iter_mut()
						.filter(|(mp3_name, _)|
							sel_mp3s.iter()
								.find(|sel_mp3| *sel_mp3 == *mp3_name) // filter tags whose name are selected
								.is_some(),
						)
					{
						sel_tag.set_text_by_field(field.name, &new_text)?; // set new frame value
					}
				}

				self2.save_cb.try_borrow()?
					.as_ref()
					.map_or(Ok(()), |cb| cb())?; // execute user callback

				Ok(())
			}
		});
	}
}
