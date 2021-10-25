use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::{prelude::*, self as w, gui, msg, ErrResult};

use crate::id3v2::{Tag, TextField};
use crate::ids::fields as id;
use super::WndFields;

impl WndFields {
	pub fn new(
		parent: &impl Parent, pos: w::POINT,
		horz: gui::Horz, vert: gui::Vert,
		tags_cache: Rc<RefCell<HashMap<String, Tag>>>) -> Self
	{
		use gui::{Button, CheckBox, ComboBox, Edit, Horz, Vert, WindowControl};

		let wnd = WindowControl::new_dlg(parent, id::DLG_FIELDS, pos, horz, vert, None);

		let chk_artist   = CheckBox::new_dlg(&wnd, id::CHK_ARTIST,   Horz::None, Vert::None);
		let txt_artist   = Edit::new_dlg(    &wnd, id::TXT_ARTIST,   Horz::None, Vert::None);
		let chk_title    = CheckBox::new_dlg(&wnd, id::CHK_TITLE,    Horz::None, Vert::None);
		let txt_title    = Edit::new_dlg(    &wnd, id::TXT_TITLE,    Horz::None, Vert::None);
		let chk_album    = CheckBox::new_dlg(&wnd, id::CHK_ALBUM,    Horz::None, Vert::None);
		let txt_album    = Edit::new_dlg(    &wnd, id::TXT_ALBUM,    Horz::None, Vert::None);
		let chk_track    = CheckBox::new_dlg(&wnd, id::CHK_TRACK,    Horz::None, Vert::None);
		let txt_track    = Edit::new_dlg(    &wnd, id::TXT_TRACK,    Horz::None, Vert::None);
		let chk_year     = CheckBox::new_dlg(&wnd, id::CHK_YEAR,     Horz::None, Vert::None);
		let txt_year     = Edit::new_dlg(    &wnd, id::TXT_YEAR,     Horz::None, Vert::None);
		let chk_genre    = CheckBox::new_dlg(&wnd, id::CHK_GENRE,    Horz::None, Vert::None);
		let cmb_genre    = ComboBox::new_dlg(&wnd, id::CMB_GENRE,    Horz::None, Vert::None);
		let chk_composer = CheckBox::new_dlg(&wnd, id::CHK_COMPOSER, Horz::None, Vert::None);
		let txt_composer = Edit::new_dlg(    &wnd, id::TXT_COMPOSER, Horz::None, Vert::None);
		let chk_comment  = CheckBox::new_dlg(&wnd, id::CHK_COMMENT,  Horz::None, Vert::None);
		let txt_comment  = Edit::new_dlg(    &wnd, id::TXT_COMMENT,  Horz::None, Vert::None);
		let btn_save     = Button::new_dlg(  &wnd, id::BTN_SAVE,     Horz::None, Vert::None);

		let fields = vec![
			(TextField::Artist,   chk_artist.clone(),   txt_artist.as_native_control()),
			(TextField::Title,    chk_title.clone(),    txt_title.as_native_control()),
			(TextField::Album,    chk_album.clone(),    txt_album.as_native_control()),
			(TextField::Track,    chk_track.clone(),    txt_track.as_native_control()),
			(TextField::Year,     chk_year.clone(),     txt_year.as_native_control()),
			(TextField::Genre,    chk_genre.clone(),    chk_genre.as_native_control()),
			(TextField::Composer, chk_composer.clone(), chk_composer.as_native_control()),
			(TextField::Comment,  chk_comment.clone(),  chk_comment.as_native_control()),
		];

		let new_self = Self {
			wnd,
			chk_artist, txt_artist,
			chk_title, txt_title,
			chk_album, txt_album,
			chk_track, txt_track,
			chk_year, txt_year,
			chk_genre, cmb_genre,
			chk_composer, txt_composer,
			chk_comment, txt_comment, btn_save,
			fields,
			tags_cache,
			sel_files: Rc::new(RefCell::new(Vec::default())),
			save_cb: Rc::new(RefCell::new(None)),
		};
		new_self._events();
		new_self
	}

	pub fn on_save<F>(&self, callback: F)
		where F: Fn() -> ErrResult<()> + 'static,
	{
		*self.save_cb.borrow_mut() = Some(Box::new(callback));
	}

	pub fn show_text_fields(&self, sel_files: Vec<String>) -> ErrResult<()> {
		self.sel_files.replace(sel_files); // keep the list of selected files

		let tags_cache = self.tags_cache.try_borrow()?;
		let sel_tags = {
			let sel_files = self.sel_files.try_borrow()?;
			tags_cache.iter()
				.filter(|(file_name, _)| {
					sel_files.iter()
						.find(|sel_file| *sel_file == *file_name) // find the tag file among the sel files
						.is_some()
				})
				.collect::<Vec<_>>()
		};

		for (field, chk, txt) in self.fields.iter() {
			let s;

			if let Some(field) = Tag::is_uniform_text_field(&sel_tags, *field)? {
				chk.set_check_state_and_trigger(gui::CheckState::Checked)?;
				s = w::WString::from_str(field);
			} else {
				chk.set_check_state_and_trigger(gui::CheckState::Unchecked)?;
				s = w::WString::from_str("");
			}

			// Works for both Edit and ComboBox.
			txt.hwnd().SendMessage(msg::wm::SetText { text: unsafe { s.as_ptr() } });
		}

		self._update_after_check()
	}
}
