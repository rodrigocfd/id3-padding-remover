use winsafe::{prelude::*, gui, ErrResult};

use crate::id3v2::TextField;
use super::WndFields;

impl WndFields {
	pub(super) fn _fields(&self) -> [(&gui::CheckBox, &dyn Child, TextField); 8] {
		[
			(&self.chk_artist,   &self.txt_artist   as _, TextField::Artist),
			(&self.chk_title,    &self.txt_title    as _, TextField::Title),
			(&self.chk_album,    &self.txt_album    as _, TextField::Album),
			(&self.chk_track,    &self.txt_track    as _, TextField::Track),
			(&self.chk_year,     &self.txt_year     as _, TextField::Year),
			(&self.chk_genre,    &self.cmb_genre    as _, TextField::Genre),
			(&self.chk_composer, &self.txt_composer as _, TextField::Composer),
			(&self.chk_comment,  &self.txt_comment  as _, TextField::Comment),
		]
	}

	pub(super) fn _update_after_check(&self) -> ErrResult<()> {
		for (chk, txt, _) in self._fields().iter() {
			txt.hwnd().EnableWindow(chk.is_checked()); // enable field if its checkbox is on
		}

		let sel_files_count = self.sel_files.try_borrow()?.len();
		let at_least_1_check = self._fields()
			.iter().find(|(chk, _, _)| chk.is_checked()).is_some();

		self.btn_save.hwnd().EnableWindow(sel_files_count > 0 && at_least_1_check);
		if at_least_1_check {
			self.btn_save.hwnd().SetWindowText(&format!("&Save ({})", sel_files_count))
		} else {
			self.btn_save.hwnd().SetWindowText("&Save")
		}?;

		Ok(())
	}
}
