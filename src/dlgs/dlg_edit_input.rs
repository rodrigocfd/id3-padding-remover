use std::sync::Arc;
use winsafe::{self as w, gui, prelude::*};

use crate::{id3v2, ids};

/// Encapsulates a CheckBox and an Edit for a field.
///
/// The notable exception is the genre, which is a ComboBox instead of an Edit.
#[derive(Clone)]
pub struct Input {
	pub(super) name4: String,
	pub(super) chk: gui::CheckBox,
	pub(super) txt: Arc<dyn GuiControl>,
}

impl Input {
	// Only receives the CheckBox ID because the Edit ID is always the next one.
	pub(super) fn new(parent: &(impl GuiParent + 'static), name4: &str, chk_id: u16) -> Self {
		let no_res = (gui::Horz::None, gui::Vert::None);

		let name4 = name4.to_owned();
		let chk = gui::CheckBox::new_dlg(parent, chk_id, no_res);
		let txt: Arc<dyn GuiControl> = if chk_id == ids::CHK_GENRE {
			Arc::new(gui::ComboBox::new_dlg(parent, chk_id + 1, no_res)) // genre is a ComboBox
		} else {
			Arc::new(gui::Edit::new_dlg(parent, chk_id + 1, no_res)) // all other frames are Edit
		};

		Self { name4, chk, txt }
	}

	pub(super) fn set_text_if_equal_in_tags(&self, sel_tags: &[id3v2::Tag]) -> w::AnyResult<()> {
		if id3v2::equal_frame_across_all_tags(&self.name4, sel_tags) {
			let user_text = sel_tags[0]
				.frame_by_name4(&self.name4)
				.unwrap()
				.as_editable_str()?;
			self.txt.hwnd().SetWindowText(&user_text)?;
			self.chk.set_check_and_trigger(true)?;
		}
		Ok(())
	}
}
