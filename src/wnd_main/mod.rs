use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::gui;

use crate::id3v2::Tag;
use crate::wnd_fields::WndFields;

mod wnd_main_events;
mod wnd_main_funcs;
mod wnd_main_menu;
mod wnd_main_privs;

#[derive(Clone)]
pub struct WndMain {
	wnd:        gui::WindowMain,
	lst_files:  gui::ListView,
	wnd_fields: WndFields,
	lst_frames: gui::ListView,
	tags_cache: Rc<RefCell<HashMap<String, Rc<RefCell<Tag>>>>>,
	app_name:   String,
}

pub enum PreDelete { Yes, No }
