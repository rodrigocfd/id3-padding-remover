use winsafe::{prelude::*, self as w, path};

use super::WndFields;

impl WndFields {
	pub(super) fn _events(&self) {
		self.wnd.on().wm_init_dialog({
			let cmb_genre = self.cmb_genre.clone();
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

		for (_, chk, _) in self.fields.iter() {
			chk.on().bn_clicked({ // add event on each checkbox
				let self2 = self.clone();
				move || self2._update_after_check()
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
