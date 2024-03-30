use winsafe::gui;

use crate::id3v2;

mod wnd_main_ctor;
mod wnd_main_funcs;
mod wnd_main_list;
mod wnd_main_wm;

#[derive(Clone)]
pub struct WndMain {
	wnd:         gui::WindowMain,
	lst_files:   gui::ListView<id3v2::Tag>,
	lst_files_h: gui::Header,
}
