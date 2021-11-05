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
				let genres = { // read genres from TXT
					let path = format!("{}\\genres.txt", path::exe_path()?);
					let fin = w::FileMapped::open(&path, w::FileAccess::ExistingReadOnly)?;
					w::WString::parse_str(fin.as_slice())?.to_string()
				};

				cmb_genre.items().add(&genres.lines().collect::<Vec<_>>())?;
				Ok(true)
			}
		});

		for field in self.fields.iter() {

			field.chk.on().bn_clicked({ // add event on each checkbox
				let self2 = self.clone();
				let field = field.clone();
				move || {
					self2._update_after_check()?;

					if field.chk.is_checked() {
						field.txt.focus()?;
					}

					Ok(())
				}
			});

		}

		self.btn_save.on().bn_clicked({
			let self2 = self.clone();
			move || {




				self2.save_cb.borrow_mut()
					.as_mut()
					.map_or(Ok(()), |cb| cb())
			}
		});
	}
}
