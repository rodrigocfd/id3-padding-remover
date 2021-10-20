use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::{self as w, co, gui, ErrResult};

use crate::id3v2::{FrameData, Tag};
use crate::ids::modify as id;
use crate::util;
use super::WndModify;

impl WndModify {
	pub fn new(parent: &impl gui::Parent,
		tags_cache: Rc<RefCell<HashMap<String, Tag>>>,
		files: Rc<Vec<String>>) -> Self
	{
		use gui::{Button, CheckBox, Horz, Vert, WindowModal};

		let wnd = WindowModal::new_dlg(parent, id::DLG_MODIFY);

		let chk_rem_padding = CheckBox::new_dlg(&wnd, id::CHK_REM_PADDING, Horz::None, Vert::None);
		let chk_rem_album   = CheckBox::new_dlg(&wnd, id::CHK_REM_ALBUM,   Horz::None, Vert::None);
		let chk_rem_rg      = CheckBox::new_dlg(&wnd, id::CHK_REM_RG,      Horz::None, Vert::None);
		let chk_prefix_year = CheckBox::new_dlg(&wnd, id::CHK_PREFIX_YEAR, Horz::None, Vert::None);

		let btn_ok     = Button::new_dlg(&wnd, id::BTN_OK, Horz::None, Vert::None);
		let btn_cancel = Button::new_dlg(&wnd, id::BTN_CANCEL, Horz::None, Vert::None);

		let new_self = Self {
			wnd,
			chk_rem_padding, chk_rem_album, chk_rem_rg, chk_prefix_year,
			btn_ok, btn_cancel,
			tags_cache, files,
		};
		new_self._events();
		new_self
	}

	pub fn show(&self) -> w::WinResult<i32> {
		self.wnd.show_modal()
	}

	pub(super) fn _enable_disable_rem_padding(&self) -> ErrResult<()> {
		// "Remove padding" checkbox will be disabled?
		let will_disable = self.chk_rem_album.is_checked()
			|| self.chk_rem_rg.is_checked()
			|| self.chk_prefix_year.is_checked();

		if will_disable { // padding removal is then always performed
			self.chk_rem_padding.set_check_state(gui::CheckState::Checked);
		}
		self.chk_rem_padding.hwnd().EnableWindow(!will_disable);

		// If won't removing padding, there's nothing to do, so we can't run.
		self.btn_ok.hwnd().EnableWindow(self.chk_rem_padding.is_checked());
		Ok(())
	}

	pub(super) fn _remove_replay_gain(&self, tag: &mut Tag) {
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

	pub(super) fn _prefix_year(&self, tag: &mut Tag, file: &str) -> ErrResult<()> {
		let frames = tag.frames_mut();

		let year = match frames.iter()
			.find(|f| f.name4() == "TYER")
			.ok_or_else(|| format!("File: {}\n\nYear frame not found.", file))?
			.data() {
			FrameData::Text(text) => text.clone(),
			_ => return Err(format!("File: {}\n\nYear frame has the wrong data type.", file).into()),
		};

		let album = match frames.iter_mut()
			.find(|f| f.name4() == "TALB")
			.ok_or_else(|| format!("File: {}\n\nAlbum frame not found.", file))?
			.data_mut() {
			FrameData::Text(text) => text,
			_ => return Err(format!("File: {}\n\nAlbum frame not found.", file).into()),
		};

		if album.starts_with(&year) {
			if util::prompt::ok_cancel(self.wnd.hwnd(), "Dubious data", None,
				&format!("File:\n{}\n\n\
					Album appears to already have the year prefix:\n{}\n\n\
					Continue anyway?",
					file, album))? != co::DLGID::OK
			{
				return Ok(()); // skip processing
			}
		}
		*album = format!("{} {}", year, album); // update album text

		Ok(())
	}
}
