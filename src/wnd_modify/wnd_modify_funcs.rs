use std::cell::RefCell;
use std::collections::HashMap;
use std::error::Error;
use std::rc::Rc;
use winsafe::{self as w, co, gui};

use crate::id3v2::{FrameData, Tag};
use crate::ids::{APP_TITLE, modify as id};
use super::WndModify;

impl WndModify {
	pub fn new(parent: &dyn gui::Parent,
		tags_cache: Rc<RefCell<HashMap<String, Tag>>>,
		files: Rc<Vec<String>>) -> Self
	{
		let wnd = gui::WindowModal::new_dlg(parent, id::DLG_MODIFY);

		let chk_rem_padding = gui::CheckBox::new_dlg(&wnd, id::CHK_REM_PADDING);
		let chk_rem_album   = gui::CheckBox::new_dlg(&wnd, id::CHK_REM_ALBUM);
		let chk_rem_rg      = gui::CheckBox::new_dlg(&wnd, id::CHK_REM_RG);
		let chk_prefix_year = gui::CheckBox::new_dlg(&wnd, id::CHK_PREFIX_YEAR);

		let btn_ok     = gui::Button::new_dlg(&wnd, id::BTN_OK);
		let btn_cancel = gui::Button::new_dlg(&wnd, id::BTN_CANCEL);

		let new_self = Self {
			wnd,
			chk_rem_padding, chk_rem_album, chk_rem_rg, chk_prefix_year,
			btn_ok, btn_cancel,
			tags_cache, files,
		};
		new_self.events();
		new_self
	}

	pub fn show(&self) {
		self.wnd.show_modal().unwrap();
	}

	pub(super) fn enable_disable_rem_padding(&self) {
		let will_disable = self.chk_rem_album.is_checked()
			|| self.chk_rem_rg.is_checked()
			|| self.chk_prefix_year.is_checked();

		if will_disable {
			self.chk_rem_padding.set_check(true);
		}
		self.chk_rem_padding.hwnd().EnableWindow(!will_disable);
	}

	pub(super) fn remove_replay_gain(&self, tag: &mut Tag) {
		tag.frames_mut().retain(|f| {
			if f.name4() == "TXXX" {
				if let FrameData::MultiText(texts) = f.data() {
					if texts[0].starts_with("replaygain_") {
						return false;
					}
				}
			}
			true
		});
	}

	pub(super) fn prefix_year(&self, tag: &mut Tag, file: &str) -> Result<(), Box<dyn Error>> {
		let frames = tag.frames_mut();

		let year = if let Some(year_frame) = frames.iter().find(|f| f.name4() == "TYER") {
			if let FrameData::Text(text) = year_frame.data() {
				text.clone()
			} else {
				return Err(format!("File: {}\n\nYear frame has the wrong data type.", file).into())
			}
		} else {
			return Err(format!("File: {}\n\nYear frame not found.", file).into())
		};

		let album = if let Some(album_frame) = frames.iter_mut().find(|f| f.name4() == "TALB") {
			if let FrameData::Text(text) = album_frame.data_mut() {
				text
			} else {
				return Err(format!("File: {}\n\nAlbum frame has the wrong data type.", file).into())
			}
		} else {
			return Err(format!("File: {}\n\nAlbum frame not found.", file).into())
		};

		if album.starts_with(&year) {
			if self.wnd.hwnd().TaskDialog(None, Some(APP_TITLE),
				Some("Dubious data"),
				Some(&format!("File:\n{}\n\n\
					Album appears to have the year prefix:\n{}.\n\n\
					Continue anyway?", file, album)),
				co::TDCBF::OK | co::TDCBF::CANCEL,
				w::IdTdicon::Tdicon(co::TD_ICON::WARNING)).unwrap() != co::DLGID::OK
			{
				return Ok(()); // skip processing
			}
		}
		*album = format!("{} {}", year, album); // update album text

		Ok(())
	}
}
