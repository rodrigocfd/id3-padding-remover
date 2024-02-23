use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::gui;

use crate::id3v2;

mod wnd_edit_ctor;
mod wnd_edit_wm;

#[derive(Clone)]
pub struct WndEdit {
	wnd:        gui::WindowModal,
	btn_ok:     gui::Button,
	btn_cancel: gui::Button,

	/// Each tag is indexed by its file path.
	all_tags: Rc<RefCell<HashMap<String, id3v2::Tag>>>,
	selected_paths: Vec<String>,
}
