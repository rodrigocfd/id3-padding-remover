use std::cell::{Cell, RefCell};
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{gui, prelude::*};

use crate::{id3v2, wnd_picture::WndPicture};

mod wnd_edit_ctor;
mod wnd_edit_wm;

#[allow(dead_code)]
#[derive(Clone)]
pub struct WndEdit {
	wnd:          gui::WindowModal,
	btn_ok:       gui::Button,
	btn_cancel:   gui::Button,
	field_packs:  Rc<RefCell<Vec<FieldPack>>>,
	wnd_pic:      WndPicture,
	lst_frames:   gui::ListView,
	sel_tags:     Vec<Rc<RefCell<id3v2::Tag>>>,
	modal_return: Rc<Cell<bool>>,
}

trait ChildFocus: GuiWindowText + GuiChildFocus {}
impl ChildFocus for gui::ComboBox {}
impl ChildFocus for gui::Edit {}

/// Known tag field identifier, checkbox and textbox.
#[derive(Clone)]
struct FieldPack {
	field: id3v2::Field,
	chk:   gui::CheckBox,
	txt:   Arc<dyn ChildFocus>,
}
