use super::WndFields;

impl WndFields {
	pub(super) fn events(&self) {
		self.chk_artist.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.txt_artist.hwnd().EnableWindow(self2.chk_artist.is_checked()); Ok(()) }
		});
		self.chk_title.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.txt_title.hwnd().EnableWindow(self2.chk_title.is_checked()); Ok(()) }
		});
		self.chk_album.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.txt_album.hwnd().EnableWindow(self2.chk_album.is_checked()); Ok(()) }
		});
		self.chk_track.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.txt_track.hwnd().EnableWindow(self2.chk_track.is_checked()); Ok(()) }
		});
		self.chk_date.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.txt_date.hwnd().EnableWindow(self2.chk_date.is_checked()); Ok(()) }
		});
		self.chk_genre.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.cmb_genre.hwnd().EnableWindow(self2.chk_genre.is_checked()); Ok(()) }
		});
		self.chk_composer.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.txt_composer.hwnd().EnableWindow(self2.chk_composer.is_checked()); Ok(()) }
		});
		self.chk_comment.on().bn_clicked({
			let self2 = self.clone();
			move || { self2.txt_comment.hwnd().EnableWindow(self2.chk_comment.is_checked()); Ok(()) }
		});
	}
}
