use std::cell::RefCell;
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{gui, prelude::*};

use crate::id3v2;

mod wnd_edit_ctor;
mod wnd_edit_wm;

#[derive(Clone)]
pub struct WndEdit {
	wnd:         gui::WindowModal,
	btn_ok:      gui::Button,
	btn_cancel:  gui::Button,
	field_packs: Rc<RefCell<Vec<FieldPack>>>,

	/// Each tag is indexed by its file path.
	all_tags:       Rc<RefCell<Vec<id3v2::PathAndTag>>>,
	selected_paths: Vec<String>,
}

/// Known tag field identifier, checkbox and textbox.
#[derive(Clone)]
struct FieldPack {
	field: id3v2::Field,
	chk:   gui::CheckBox,
	txt:   Arc<dyn GuiWindowText>,
}
