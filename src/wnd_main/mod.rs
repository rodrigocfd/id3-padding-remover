use std::cell::Cell;
use std::rc::Rc;
use winsafe::gui;

use crate::id3v2;

mod wnd_main_ctor;
mod wnd_main_funcs;
mod wnd_main_lvn;
mod wnd_main_wm;

#[derive(Clone)]
pub struct WndMain {
	wnd:          gui::WindowMain,
	lst_files:    gui::ListView<id3v2::Tag>,
	lst_files_h:  gui::Header,
	cur_sort_col: Rc<Cell<(u32, bool)>>, // index, reversed
}

const LIST_COLS: &[(&str, u32, &str)] = &[
	("File", 380, ""),
	("Pad", 50, ""),
	("Art", 30, ""),
	("Artist", 160, "TPE1"),
	("Title", 180, "TIT2"),
	("Album", 180, "TALB"),
	("T#", 30, "TRCK"),
	("Year", 40, "TYER"),
	("Genre", 100, "TCON"),
	("Comment", 80, "COMM"),
];
