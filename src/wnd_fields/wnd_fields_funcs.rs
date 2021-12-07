use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::{Arc, Mutex};
use winsafe::{prelude::*, self as w, gui};

use crate::{id3v2, wnd_picture::WndPicture};
use super::{ids, Field, WndFields};

impl WndFields {
	pub fn new(
		parent: &impl Parent,
		tags_cache: Arc<Mutex<HashMap<String, id3v2::Tag>>>,
		pos: w::POINT,
		resize_behavior: (gui::Horz, gui::Vert)) -> Self
	{
		let wnd = gui::WindowControl::new_dlg(parent, ids::DLG_FIELDS, pos, resize_behavior, None);

		let hv_none = (gui::Horz::None, gui::Vert::None);
		let fields = [
			("TPE1", ids::CHK_ARTIST,      ids::TXT_ARTIST),
			("TIT2", ids::CHK_TITLE,       ids::TXT_TITLE),
			("TIT3", ids::CHK_SUBTITLE,    ids::TXT_SUBTITLE),
			("TALB", ids::CHK_ALBUM,       ids::TXT_ALBUM),
			("TRCK", ids::CHK_TRACK,       ids::TXT_TRACK),
			("TYER", ids::CHK_YEAR,        ids::TXT_YEAR),
			("TCON", ids::CHK_GENRE,       ids::CMB_GENRE),
			("TCOM", ids::CHK_COMPOSER,    ids::TXT_COMPOSER),
			("TEXT", ids::CHK_LYRICIST,    ids::TXT_LYRICIST),
			("TOPE", ids::CHK_ORIG_ARTIST, ids::TXT_ORIG_ARTIST),
			("TOAL", ids::CHK_ORIG_ALBUM,  ids::TXT_ORIG_ALBUM),
			("TORY", ids::CHK_ORIG_YEAR,   ids::TXT_ORIG_YEAR),
			("TPE3", ids::CHK_PERFORMER,   ids::TXT_PERFORMER),
			("COMM", ids::CHK_COMMENT,     ids::TXT_COMMENT),
		].map(|(name4, id_chk, id_txt)| Field { // dynamically build all the field structs
			name4,
			chk: gui::CheckBox::new_dlg(&wnd, id_chk, hv_none),
			txt: if id_txt == ids::CMB_GENRE {
				Arc::new(gui::ComboBox::new_dlg(&wnd, id_txt, hv_none))
			} else {
				Arc::new(gui::Edit::new_dlg(&wnd, id_txt, hv_none))
			},
		}).to_vec();

		let wnd_picture      = WndPicture::new(&wnd, w::POINT::new(160, 66), w::SIZE::new(30, 30), hv_none);
		let btn_clear_checks = gui::Button::new_dlg(&wnd, ids::BTN_CLEARCHECKS, hv_none);
		let btn_save         = gui::Button::new_dlg(&wnd, ids::BTN_SAVE, hv_none);

		let new_self = Self {
			wnd, fields, wnd_picture, btn_clear_checks, btn_save,
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

		for field in self.fields.iter() { // text fields
			let (chk_state, text) = if id3v2::Tag::same_frame_value(&sel_tags, field.name4)? {
				(gui::CheckState::Checked, sel_tags[0].text_of_frame(field.name4)?.unwrap())
			} else {
				(gui::CheckState::Unchecked, "")
			};

			field.chk.hwnd().EnableWindow(!sel_mp3s.is_empty()); // if zero MP3s selected, disable checkboxes
			field.chk.set_check_state(chk_state);
			field.txt.set_text(text)?;
			field.txt.hwnd().EnableWindow(chk_state == gui::CheckState::Checked);
		}

		let same_pic = id3v2::Tag::same_frame_value(&sel_tags, "APIC")?;
		if same_pic {
			let pic_frame = sel_tags[0].frame_by_name4("APIC").unwrap();
			if let id3v2::FrameData::Picture(pic) = pic_frame.data() {
				self.wnd_picture.feed(Some(&pic.pic_bytes))?;
			}
		} else {
			self.wnd_picture.feed(None)?;
		}
		self.wnd_picture.enable(same_pic);


		*self.sel_mp3s.try_borrow_mut()? = sel_mp3s; // keep selected files
		self._enable_buttons_if_at_least_one_checked();
		Ok(())
	}
}
