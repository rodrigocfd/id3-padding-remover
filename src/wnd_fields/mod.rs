use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{self as w, gui};

use crate::id3v2::{Tag, TextField};

mod wnd_fields_events;
mod wnd_fields_funcs;
mod wnd_fields_privs;

#[derive(Clone)]
pub struct WndFields {
	wnd: gui::WindowControl,

	chk_artist:   gui::CheckBox,
	txt_artist:   gui::Edit,
	chk_title:    gui::CheckBox,
	txt_title:    gui::Edit,
	chk_album:    gui::CheckBox,
	txt_album:    gui::Edit,
	chk_track:    gui::CheckBox,
	txt_track:    gui::Edit,
	chk_year:     gui::CheckBox,
	txt_year:     gui::Edit,
	chk_genre:    gui::CheckBox,
	cmb_genre:    gui::ComboBox,
	chk_composer: gui::CheckBox,
	txt_composer: gui::Edit,
	chk_comment:  gui::CheckBox,
	txt_comment:  gui::Edit,
	btn_save:     gui::Button,

	fields:     Vec<(TextField, gui::CheckBox, Arc<dyn gui::NativeControl>)>,
	tags_cache: Rc<RefCell<HashMap<String, Tag>>>,
	sel_files:  Rc<RefCell<Vec<String>>>,
	save_cb:    Rc<RefCell<Option<Box<dyn Fn() -> w::ErrResult<()>>>>>,
}
