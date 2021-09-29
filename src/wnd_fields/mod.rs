use winsafe::{self as w, gui};

mod wnd_fields_events;
mod wnd_fields_funcs;

#[derive(Clone)]
pub struct WndFields {
	wnd: gui::WindowControl,

	chk_artist:   gui::CheckBox,
	txt_artist:   gui::Edit,
	chk_title:    gui::CheckBox,
	txt_title:    gui::Edit,
	chk_album:    gui::CheckBox,
	txt_album:    gui::Edit,
	chk_track:    gui::CheckBox,
	txt_track:    gui::Edit,
	chk_date:     gui::CheckBox,
	txt_date:     gui::Edit,
	chk_genre:    gui::CheckBox,
	cmb_genre:    gui::ComboBox,
	chk_composer: gui::CheckBox,
	txt_composer: gui::Edit,
	chk_comment:  gui::CheckBox,
	txt_comment:  gui::Edit,
	btn_save:     gui::Button,
}

impl gui::Child for WndFields {
	fn hwnd_ref(&self) -> &w::HWND {
		self.wnd.hwnd_ref()
	}
}
