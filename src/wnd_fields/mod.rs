use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::{Arc, Mutex};
use winsafe::{self as w, gui};

use crate::id3v2;

mod ids;
mod wnd_fields_events;
mod wnd_fields_funcs;
mod wnd_fields_privs;

/// Control with lots of checkboxes and textboxes.
#[derive(Clone)]
pub struct WndFields {
	wnd:              gui::WindowControl,
	fields:           Vec<Field>,
	btn_clear_checks: gui::Button,
	btn_save:         gui::Button,
	tags_cache:       Arc<Mutex<HashMap<String, id3v2::Tag>>>,
	sel_mp3s:         Rc<RefCell<Vec<String>>>,
	save_cb:          Rc<RefCell<Option<Box<dyn Fn() -> w::ErrResult<()>>>>>,
}

#[derive(Clone)]
struct Field {
	name: id3v2::FieldName,
	chk:  gui::CheckBox,
	txt:  Arc<dyn TxtCtrl>,
}

trait TxtCtrl: w::prelude::TextControl + w::prelude::FocusControl {}
impl TxtCtrl for gui::Edit {}
impl TxtCtrl for gui::ComboBox {}
