use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::{Arc, Mutex};
use winsafe::{prelude::*, self as w, gui};

use crate::id3v2;
use super::{ids, Field, WndFields};

impl WndFields {
	pub fn new(
		parent: &impl Parent,
		tags_cache: Arc<Mutex<HashMap<String, id3v2::Tag>>>,
		pos: w::POINT,
		resize_behavior: (gui::Horz, gui::Vert)) -> Self
	{
		let wnd = gui::WindowControl::new_dlg(parent, ids::DLG_FIELDS, pos, resize_behavior, None);

		let none2 = (gui::Horz::None, gui::Vert::None);
		use id3v2::FieldName::*;
		let fields = [
			(Artist,     ids::CHK_ARTIST,      ids::TXT_ARTIST),
			(Title,      ids::CHK_TITLE,       ids::TXT_TITLE),
			(Subtitle,   ids::CHK_SUBTITLE,    ids::TXT_SUBTITLE),
			(Album,      ids::CHK_ALBUM,       ids::TXT_ALBUM),
			(Track,      ids::CHK_TRACK,       ids::TXT_TRACK),
			(Year,       ids::CHK_YEAR,        ids::TXT_YEAR),
			(Genre,      ids::CHK_GENRE,       ids::CMB_GENRE),
			(Composer,   ids::CHK_COMPOSER,    ids::TXT_COMPOSER),
			(Lyricist,   ids::CHK_LYRICIST,    ids::TXT_LYRICIST),
			(OrigArtist, ids::CHK_ORIG_ARTIST, ids::TXT_ORIG_ARTIST),
			(OrigAlbum,  ids::CHK_ORIG_ALBUM,  ids::TXT_ORIG_ALBUM),
			(OrigYear,   ids::CHK_ORIG_YEAR,   ids::TXT_ORIG_YEAR),
			(Performer,  ids::CHK_PERFORMER,   ids::TXT_PERFORMER),
			(Comment,    ids::CHK_COMMENT,     ids::TXT_COMMENT),
		].map(|(name, idchk, idtxt)| Field { // dynamically build all the frame fields
			name,
			chk: gui::CheckBox::new_dlg(&wnd, idchk, none2),
			txt: if idtxt == ids::CMB_GENRE {
				Arc::new(gui::ComboBox::new_dlg(&wnd, idtxt, none2))
			} else {
				Arc::new(gui::Edit::new_dlg(&wnd, idtxt, none2))
			},
		}).to_vec();

		let btn_clear_checks = gui::Button::new_dlg(&wnd, ids::BTN_CLEARCHECKS, none2);
		let btn_save         = gui::Button::new_dlg(&wnd, ids::BTN_SAVE, none2);

		let new_self = Self {
			wnd, fields, btn_clear_checks, btn_save,
			tags_cache,
			sel_mp3s: Rc::new(RefCell::new(Vec::default())),
			save_cb:   Rc::new(RefCell::new(None)),
		};
		new_self._events();
		new_self
	}

	pub fn on_save<F>(&self, callback: F) -> w::ErrResult<()>
		where F: Fn() -> w::ErrResult<()> + 'static,
	{
		*self.save_cb.try_borrow_mut()? = Some(Box::new(callback)); // store callback
		Ok(())
	}

	pub fn feed(&self, sel_mp3s: Vec<String>) -> w::ErrResult<()> {
		let tags_cache = self.tags_cache.lock().unwrap();
		let sel_tags = tags_cache.iter()
			.filter(|(file_name, _)|
				sel_mp3s.iter()
					.find(|sel_file| *sel_file == *file_name)
					.is_some(),
			)
			.map(|(_, tag)| tag)
			.collect::<Vec<_>>();

		for field in self.fields.iter() {
			let (check_state, s) = match id3v2::Tag::same_field_value(&sel_tags, field.name)? {
				Some(s) => (gui::CheckState::Checked, w::WString::from_str(&s)),
				None => (gui::CheckState::Unchecked, w::WString::from_str("")),
			};

			field.chk.hwnd().EnableWindow(!sel_mp3s.is_empty()); // if zero MP3s selected, disable checkboxes
			field.chk.set_check_state(check_state);
			field.txt.set_text(&s.to_string())?;
			field.txt.hwnd().EnableWindow(check_state == gui::CheckState::Checked);
		}

		*self.sel_mp3s.try_borrow_mut()? = sel_mp3s; // keep selected files
		self._enable_buttons_if_at_least_one_checked();
		Ok(())
	}
}
