use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::gui;

use crate::id3v2::Tag;

mod wnd_main_events;
mod wnd_main_funcs;
mod wnd_main_menu;

#[derive(Clone)]
pub struct WndMain {
	wnd:        gui::WindowMain,
	lst_files:  gui::ListView,
	lst_frames: gui::ListView,
	resizer:    gui::Resizer,
	tags:       Rc<RefCell<HashMap<String, Tag>>>,
}
