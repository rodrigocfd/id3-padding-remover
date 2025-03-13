use std::sync::Arc;
use winsafe::{gui, prelude::*};

/// Known tag field identifier, checkbox and textbox.
#[derive(Clone)]
pub struct Input {
	pub(super) name4: String,
	pub(super) chk: gui::CheckBox,
	pub(super) txt: Arc<dyn GuiControl>,
}

impl Input {
	pub(super) fn new_edit(name4: &str, parent: &(impl GuiParent + 'static), chk_id: u16) -> Self {
		let none2 = (gui::Horz::None, gui::Vert::None);
		Self {
			name4: name4.to_owned(),
			chk: gui::CheckBox::new_dlg(parent, chk_id, none2),
			txt: Arc::new(gui::Edit::new_dlg(parent, chk_id + 1, none2)),
		}
	}
}
