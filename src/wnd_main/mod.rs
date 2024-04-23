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
	("File", 400, ""),
	("Pad", 50, ""),
	("Art", 30, ""),
	("RG", 30, ""),
	("Artist", 90, "TPE1"),
	("T#", 30, "TRCK"),
	("Title", 100, "TIT2"),
	("Album", 100, "TALB"),
	("Year", 40, "TYER"),
	("Genre", 90, "TCON"),
	("Performer", 80, "TPE3"),
	("Composer", 80, "TCOM"),
	("Lyricist", 80, "TEXT"),
	("Comment", 80, "COMM"),
];
