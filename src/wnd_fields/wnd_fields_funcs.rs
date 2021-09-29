use winsafe::{self as w, gui};

use crate::ids::fields as id;
use super::WndFields;

impl WndFields {
	pub fn new(parent: &dyn gui::Parent, pos: w::POINT) -> Self {
		let wnd = gui::WindowControl::new_dlg(parent, id::DLG_FIELDS, pos, None);

		let chk_artist = gui::CheckBox::new_dlg(&wnd, id::CHK_ARTIST);
		let txt_artist = gui::Edit::new_dlg(&wnd, id::TXT_ARTIST);
		let chk_title = gui::CheckBox::new_dlg(&wnd, id::CHK_TITLE);
		let txt_title = gui::Edit::new_dlg(&wnd, id::TXT_TITLE);
		let chk_album = gui::CheckBox::new_dlg(&wnd, id::CHK_ALBUM);
		let txt_album = gui::Edit::new_dlg(&wnd, id::TXT_ALBUM);
		let chk_track = gui::CheckBox::new_dlg(&wnd, id::CHK_TRACK);
		let txt_track = gui::Edit::new_dlg(&wnd, id::TXT_TRACK);
		let chk_date = gui::CheckBox::new_dlg(&wnd, id::CHK_DATE);
		let txt_date = gui::Edit::new_dlg(&wnd, id::TXT_DATE);
		let chk_genre = gui::CheckBox::new_dlg(&wnd, id::CHK_GENRE);
		let cmb_genre = gui::ComboBox::new_dlg(&wnd, id::CMB_GENRE);
		let chk_composer = gui::CheckBox::new_dlg(&wnd, id::CHK_COMPOSER);
		let txt_composer = gui::Edit::new_dlg(&wnd, id::TXT_COMPOSER);
		let chk_comment = gui::CheckBox::new_dlg(&wnd, id::CHK_COMMENT);
		let txt_comment = gui::Edit::new_dlg(&wnd, id::TXT_COMMENT);
		let btn_save = gui::Button::new_dlg(&wnd, id::BTN_SAVE);

		let new_self = Self {
			wnd,
			chk_artist, txt_artist,
			chk_title, txt_title,
			chk_album, txt_album,
			chk_track, txt_track,
			chk_date, txt_date,
			chk_genre, cmb_genre,
			chk_composer, txt_composer,
			chk_comment, txt_comment, btn_save,
		};
		new_self.events();
		new_self
	}
}
