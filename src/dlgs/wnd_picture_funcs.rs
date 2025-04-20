use winsafe::{self as w, prelude::*};

use super::WndPicture;
use crate::{id3v2, msgbox};

impl WndPicture {
	pub(super) fn load_picture(&self, sel_tags: &[id3v2::Tag]) -> w::AnyResult<()> {
		if !id3v2::equal_frame_across_all_tags("APIC", sel_tags) {
			return Ok(());
		}

		match &sel_tags[0].frame_by_name4("APIC").unwrap().body() {
			id3v2::Body::Picture(pic_body) => {
				let hparent = self.wnd.hwnd().GetParent()?;

				let stream = match w::SHCreateMemStream(&pic_body.data) {
					Err(err) => {
						msgbox::err(
							&hparent,
							"Cannot load picture stream",
							None,
							&format!("SHCreateMemStream error:\n{}", &err.to_string()),
						)?;
						self.pic_err.set(Some(err)); // store the error
						return Ok(());
					},
					Ok(stream) => stream,
				};

				let pic_obj = match w::OleLoadPicture(&stream, None, true) {
					Err(err) => {
						msgbox::err(
							&hparent,
							"Cannot parse picture",
							None,
							&format!("OleLoadPicture error:\n{}", &err.to_string()),
						)?;
						self.pic_err.set(Some(err)); // store the error
						return Ok(());
					},
					Ok(pic_obj) => pic_obj,
				};

				*self.pic.try_borrow_mut()? = Some(pic_obj);
				self.wnd.hwnd().InvalidateRect(None, true)?;
			},
			_ => return Err("APIC body is not picture.".into()), // should never happen
		}

		Ok(())
	}

	pub(super) fn unload_picture(&self) -> w::AnyResult<()> {
		*self.pic.try_borrow_mut()? = None;
		self.wnd.hwnd().InvalidateRect(None, true)?;
		Ok(())
	}
}
