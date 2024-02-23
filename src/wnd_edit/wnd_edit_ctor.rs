use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::{self as w, prelude::*, co, gui};

use crate::{id3v2, ids};
use super::WndEdit;

impl WndEdit {
	/// Creates a new `WndEdit` object.
	#[must_use]
	pub fn new(
		parent: &impl GuiParent,
		all_tags: Rc<RefCell<HashMap<String, id3v2::Tag>>>,
		selected_paths: Vec<String>,
	) -> Self
	{
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowModal::new_dlg(parent, ids::DLG_EDIT);
		let btn_ok = gui::Button::new_dlg(&wnd, co::DLGID::OK.into(), (H::Repos, V::Repos));
		let btn_cancel = gui::Button::new_dlg(&wnd, co::DLGID::CANCEL.into(), (H::Repos, V::Repos));

		let new_self = Self { wnd, btn_ok, btn_cancel, all_tags, selected_paths };
		new_self.wm_events();
		new_self
	}

	pub fn show(&self) -> w::AnyResult<()> {
		self.wnd.show_modal()
			.map(|_| ())
	}

	/// Initializes the `WndEdit` window.
	pub(super) fn init_dialog(&self) -> w::AnyResult<bool> {

		Ok(true)
	}
}
