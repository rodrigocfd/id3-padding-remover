use winsafe::{self as w, gui, prelude::*};

use super::DlgEdit;
use crate::{id3v2, ids};

impl DlgEdit {
	pub(super) fn load_combo_genres(&self) -> w::AnyResult<()> {
		if let Some(input) = self // find the genres ComboBox among the inputs
			.inputs
			.iter()
			.find(|input| input.txt.ctrl_id() == ids::CMB_GENRE)
		{
			let cmb_genres = input.txt.as_any().downcast_ref::<gui::ComboBox>().unwrap();
			let genres = include_str!("genres.txt");
			genres
				.lines()
				.filter(|line| !line.is_empty())
				.try_for_each(|genre| cmb_genres.items().add(&[genre]))?;
		}
		Ok(())
	}

	pub(super) fn render_frames_list(&self) -> w::AnyResult<()> {
		self.lst_frames.items().delete_all()?;

		let sel_tags = self.sel_tags.try_borrow()?;
		if sel_tags.len() == 1 {
			// Editing only 1 MP3 file, render each frame.
			sel_tags[0]
				.frames()
				.iter()
				.try_for_each(|frame| -> w::AnyResult<_> {
					let text = frame.body().to_string(); // textual representation of the frame data
					self.lst_frames
						.items()
						.add(&[frame.name4(), &text], None, frame.clone())?; // store a copy of the frame in the item
					Ok(())
				})?;
		} else {
			// Editing multiple MP3 files, just display a file count.
			let text = format!("{} files...", sel_tags.len());
			self.lst_frames
				.items()
				.add(&["", &text], None, id3v2::Frame::new_empty())?;
			self.lst_frames.hwnd().EnableWindow(false);
		}
		Ok(())
	}

	pub(super) fn load_picture(&self) -> w::AnyResult<()> {
		self.wnd_pic.load_picture(&self.sel_tags.try_borrow()?)?; // ask the control to load the IPicture
		if let Some(pic_obj) = &*self.wnd_pic.pic.try_borrow()? {
			// We have a picture loaded.
			self.chk_pic.set_check(true);
			let (cx, cy) = {
				let hdc_screen = w::HWND::NULL.GetDC()?;
				hdc_screen.HiMetricToPixel(pic_obj.get_Width()?, pic_obj.get_Height()?)
			};
			self.lbl_pic
				.hwnd()
				.SetWindowText(&format!("{cx} x {cy} px"))?;
		} else {
			if self.wnd_pic.pic_err.get().is_some() {
				// There is a picture, but it failed to render.
				self.chk_pic.set_check(true);
				self.lbl_pic.hwnd().SetWindowText("(failed to load)")?;
			} else {
				// We don't have a picture loaded.
				self.chk_pic.set_check(false);
				self.lbl_pic.hwnd().SetWindowText("")?;
			}
		}
		Ok(())
	}
}
