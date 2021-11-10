use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{prelude::*, self as w, gui};

use crate::id3v2;
use crate::ids::fields as id;
use super::{Field, WndFields};

impl WndFields {
	pub fn new(
		parent: &impl Parent,
		tags_cache: Rc<RefCell<HashMap<String, id3v2::Tag>>>,
		pos: w::POINT,
		horz: gui::Horz, vert: gui::Vert) -> Self
	{
		let wnd = gui::WindowControl::new_dlg(parent, id::DLG_FIELDS, pos, horz, vert, None);

		use gui::{Horz::None as HNone, Vert::None as VNone};
		use id3v2::FieldName::*;
		let fields = [
			(Artist,   id::CHK_ARTIST,   id::TXT_ARTIST),
			(Title,    id::CHK_TITLE,    id::TXT_TITLE),
			(Album,    id::CHK_ALBUM,    id::TXT_ALBUM),
			(Track,    id::CHK_TRACK,    id::TXT_TRACK),
			(Year,     id::CHK_YEAR,     id::TXT_YEAR),
			(Genre,    id::CHK_GENRE,    id::CMB_GENRE),
			(Composer, id::CHK_COMPOSER, id::TXT_COMPOSER),
			(Comment,  id::CHK_COMMENT,  id::TXT_COMMENT),
		].map(|(name, idchk, idtxt)| Field {
			name,
			chk: gui::CheckBox::new_dlg(&wnd, idchk, HNone, VNone),
			txt: if idtxt == id::CMB_GENRE {
				Arc::new(gui::ComboBox::new_dlg(&wnd, idtxt, HNone, VNone))
			} else {
				Arc::new(gui::Edit::new_dlg(&wnd, idtxt, HNone, VNone))
			},
		}).to_vec();

		let btn_save = gui::Button::new_dlg(&wnd, id::BTN_SAVE, HNone, VNone);

		let new_self = Self {
			wnd, fields, btn_save,
			tags_cache,
			sel_files: Rc::new(RefCell::new(Vec::default())),
			save_cb:   Rc::new(RefCell::new(None)),
		};
		new_self._events();
		new_self
	}

	pub fn on_save<F>(&self, callback: F) -> w::ErrResult<()>
		where F: Fn() -> w::ErrResult<()> + 'static,
	{
		*self.save_cb.try_borrow_mut()? = Some(Box::new(callback));
		Ok(())
	}

	pub fn feed(&self, sel_files: Vec<String>) -> w::ErrResult<()> {
		let tags_cache = self.tags_cache.try_borrow()?;
		let sel_tags = tags_cache.iter()
			.filter(|(file_name, _)|
				sel_files.iter()
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

			field.chk.set_check_state(check_state);
			field.txt.set_text(&s.to_string())?;
			field.txt.hwnd().EnableWindow(check_state == gui::CheckState::Checked);
		}

		*self.sel_files.try_borrow_mut()? = sel_files; // keep selected files
		self._update_after_check()
	}
}
