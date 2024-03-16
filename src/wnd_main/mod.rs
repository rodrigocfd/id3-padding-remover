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
	wnd:         gui::WindowMain,
	lst_files:   gui::ListView,
	lst_files_h: gui::Header,

	/// Each tag is indexed by its file path.
	all_tags: Rc<RefCell<Vec<id3v2::PathAndTag>>>,
}
