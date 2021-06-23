use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::gui;

use crate::id3v2::Tag;

mod wnd_modify_events;
mod wnd_modify_funcs;

#[derive(Clone)]
pub struct WndModify {
	wnd:             gui::WindowModal,
	chk_rem_padding: gui::CheckBox,
	chk_rem_album:   gui::CheckBox,
	chk_rem_rg:      gui::CheckBox,
	chk_prefix_year: gui::CheckBox,
	btn_ok:          gui::Button,
	btn_cancel:      gui::Button,

	tags_cache: Rc<RefCell<HashMap<String, Tag>>>,
	files:      Rc<Vec<String>>,
}
