use std::cell::RefCell;
use std::rc::Rc;
use winsafe::gui;

use crate::id3v2;

mod wnd_main_ctor;
mod wnd_main_funcs;
mod wnd_main_list;
mod wnd_main_wm;

#[derive(Clone)]
pub struct WndMain {
	wnd:       gui::WindowMain,
	lst_files: gui::ListView,
	tags:      Rc<RefCell<Vec<TagInfo>>>,
}

pub struct TagInfo {
	pub mp3_path: String,
	pub tag:      id3v2::Tag,
}
