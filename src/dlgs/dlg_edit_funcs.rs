use try_iterator::prelude::*;
use winsafe::{self as w, prelude::*, co, gui};

use super::DlgEdit;

impl DlgEdit {
	pub(super) fn fill_chks_and_txts(&self) -> w::AnyResult<()> {
		let genres = include_str!("genres.txt");

		self.inputs.try_borrow()?
			.iter()
			.try_for_each(|input| {
				if input.name4 == "TCON" { // feed the genres to the combo
					let cmb = input.txt.as_any()
						.downcast_ref::<gui::ComboBox>()
						.expect("ComboBox downcast failed.");
					genres.lines()
						.filter(|line| !line.is_empty())
						.try_for_each(|genre| {
							cmb.items().add(&[genre])?;
							w::SysResult::Ok(())
						})?;
				}

				let maybe_idx_first_mp3 = self.sel_tags.iter() // index of first MP3 which has the field
					.try_position(|tag| {
						let has = tag.try_borrow()?.frame(&input.name4).is_some();
						w::AnyResult::Ok(has)
					})?;

				match maybe_idx_first_mp3 {
					Some(idx_first) => { // at least 1 MP3 has this field
						let first_tag = self.sel_tags[idx_first].try_borrow()?;
						let first_frame = first_tag.frame(&input.name4).unwrap();

						let frame_equal_in_all_mp3s = self.sel_tags.iter()
							.skip(idx_first + 1)
							.try_all(|tag| {
								let is_equal_to_1st = match tag.try_borrow()?.frame(&input.name4) {
									None => false, // this MP3 doesn't have this field
									Some(frame) => frame == first_frame,
								};
								w::AnyResult::Ok(is_equal_to_1st)
							})?;

						if frame_equal_in_all_mp3s {
							input.txt.hwnd().SetWindowText(&first_frame.body().to_string())?;
							input.chk.set_check_and_trigger(true)?;
						} else {
							input.chk.set_check_and_trigger(false)?;
						}
					},
					None => { // no MP3 has this field
						input.chk.set_check_and_trigger(false)?;
					},
				}

				w::AnyResult::Ok(())
			})
	}

	pub(super) fn fill_listview_fields(&self) -> w::AnyResult<()> {
		self.lst_frames.cols().add("Frame", gui::dpi_x(56))?;
		self.lst_frames.cols().add("Value", 1)?.set_width_to_fill()?;
		self.lst_frames.set_extended_style(true, co::LVS_EX::FULLROWSELECT | co::LVS_EX::GRIDLINES);

		if self.sel_tags.len() > 1 {
			self.lst_frames.items().add(&["", &format!("{} files...", self.sel_tags.len())], None, ())?;
		} else {
			let sel_tag = self.sel_tags[0].try_borrow()?;
			sel_tag.frames()
				.iter()
				.try_for_each(|frame| {
					self.lst_frames.items().add(&[frame.name4(), &frame.body().to_string()], None, ())?;
					w::SysResult::Ok(())
				})?;
		}

		Ok(())
	}
}
